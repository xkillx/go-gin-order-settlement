package service

import (
    "context"
    "encoding/csv"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/xkillx/go-gin-order-settlement/database/entities"
	jobrepo "github.com/xkillx/go-gin-order-settlement/modules/job/repository"
	settrepo "github.com/xkillx/go-gin-order-settlement/modules/settlement/repository"
	txrepo "github.com/xkillx/go-gin-order-settlement/modules/transaction/repository"
)

const (
	jobStatusQueued    = "QUEUED"
	jobStatusRunning   = "RUNNING"
	jobStatusCompleted = "COMPLETED"
	jobStatusCancelled = "CANCELLED"
	jobStatusFailed    = "FAILED"
)

// JobManager coordinates settlement jobs over transactions.
type JobManager struct {
	transactionRepo txrepo.TransactionRepo
	settlementRepo  settrepo.SettlementRepo
	jobRepo         jobrepo.JobRepo

	workers   int
	batchSize int

	cancelMu sync.Mutex
	cancels  map[string]context.CancelFunc
}

// NewJobManager constructs a JobManager reading WORKERS and BATCH_SIZE from env with sane defaults.
func NewJobManager(t txrepo.TransactionRepo, s settrepo.SettlementRepo, j jobrepo.JobRepo) *JobManager {
	workers := getEnvInt("WORKERS", runtime.NumCPU())
	if workers < 1 {
		workers = 1
	}
	batchSize := getEnvInt("BATCH_SIZE", 1000)
	if batchSize < 1 {
		batchSize = 1000
	}
	return &JobManager{
		transactionRepo: t,
		settlementRepo:  s,
		jobRepo:         j,
		workers:         workers,
		batchSize:       batchSize,
		cancels:         make(map[string]context.CancelFunc),
	}
}

// StartSettlementJob creates a job record and launches processing in background.
// It returns immediately with the job ID (HTTP 202 semantics up to the caller).
func (m *JobManager) StartSettlementJob(ctx context.Context, fromDate, toDate time.Time) (string, error) {
	// Count transactions for progress/estimation
	total, err := m.transactionRepo.Count(ctx, fromDate, toDate)
	if err != nil {
		return "", err
	}

	jobID := uuid.NewString()
	job := entities.Job{
		ID:       jobID,
		Status:   jobStatusQueued,
		FromDate: fromDate,
		ToDate:   toDate,
		Total:    total,
	}
	if err := m.jobRepo.Create(ctx, job); err != nil {
		return "", err
	}

	// Fire-and-forget processing goroutine
	go m.runSettlementJob(ctx, jobID, fromDate, toDate)

	return jobID, nil
}

// Cancel stops a running job if present. Returns true if a cancel was triggered.
func (m *JobManager) Cancel(jobID string) bool {
	m.cancelMu.Lock()
	defer m.cancelMu.Unlock()
	if c, ok := m.cancels[jobID]; ok {
		c()
		return true
	}
	return false
}

// internal helper type for worker -> collector communication
// count indicates how many transactions contributed to this aggregate
// so progress can be updated accurately.
type partialResult struct {
	agg   map[string]entities.Settlement
	count int
}

func (m *JobManager) runSettlementJob(parentCtx context.Context, jobID string, from, to time.Time) {
	// Update status to RUNNING
	_ = m.jobRepo.SetStatus(parentCtx, jobID, jobStatusRunning)

	// Derive cancellable context and store cancel function
	jobCtx, cancel := context.WithCancel(parentCtx)
	m.cancelMu.Lock()
	m.cancels[jobID] = cancel
	m.cancelMu.Unlock()
	defer func() {
		m.cancelMu.Lock()
		delete(m.cancels, jobID)
		m.cancelMu.Unlock()
		cancel()
	}()

	// Prepare CSV output
	outDir := "/tmp/settlements"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		m.fail(jobCtx, jobID, fmt.Errorf("mkdir: %w", err))
		return
	}
	outPath := filepath.Join(outDir, jobID+".csv")
	f, err := os.Create(outPath)
	if err != nil {
		m.fail(jobCtx, jobID, fmt.Errorf("create csv: %w", err))
		return
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"merchant_id", "date", "gross_cents", "fee_cents", "net_cents", "txn_count"})
	w.Flush()
	if err := w.Error(); err != nil {
		m.fail(jobCtx, jobID, fmt.Errorf("csv header: %w", err))
		return
	}

	// Retrieve total from DB (created in StartSettlementJob)
	job, err := m.jobRepo.Get(jobCtx, jobID)
	if err != nil {
		m.fail(jobCtx, jobID, fmt.Errorf("get job: %w", err))
		return
	}
	total := job.Total

	// Channels and concurrency setup
	batchChan := make(chan []entities.Transaction, m.workers*2)
	resultChan := make(chan partialResult, m.workers*2)
	producerErr := make(chan error, 1)

	// Start producer
	go func() {
		err := m.transactionRepo.StreamByDateRange(jobCtx, from, to, m.batchSize, batchChan)
		if err != nil {
			producerErr <- err
		}
		close(batchChan)
	}()

	// Start workers
	var wgWorkers sync.WaitGroup
	wgWorkers.Add(m.workers)
	for i := 0; i < m.workers; i++ {
		go func() {
			defer wgWorkers.Done()
			for {
				select {
				case <-jobCtx.Done():
					return
				case batch, ok := <-batchChan:
					if !ok {
						return
					}
					// Pre-check cancellation
					select {
					case <-jobCtx.Done():
						return
					default:
					}

					agg := make(map[string]entities.Settlement, len(batch))
					for _, tx := range batch {
						day := time.Date(tx.PaidAt.UTC().Year(), tx.PaidAt.UTC().Month(), tx.PaidAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
						key := tx.MerchantID + "|" + day.Format("2006-01-02")
						cur := agg[key]
						cur.MerchantID = tx.MerchantID
						cur.Date = day
						cur.GrossCents += tx.AmountCents
						cur.FeeCents += tx.FeeCents
						cur.NetCents += tx.AmountCents - tx.FeeCents
						cur.TxnCount += 1
						agg[key] = cur
					}
					// Post-check cancellation
					select {
					case <-jobCtx.Done():
						return
					case resultChan <- partialResult{agg: agg, count: len(batch)}:
					}
				}
			}
		}()
	}

	// Close resultChan once all workers are done
	go func() {
		wgWorkers.Wait()
		close(resultChan)
	}()

	// Collector: merge and periodically flush
	global := make(map[string]*entities.Settlement)
	changed := make(map[string]struct{})
	batchesSinceFlush := 0
	const flushEveryBatches = 50
	var processed int64

	flush := func(force bool) error {
		if len(changed) == 0 && !force {
			return nil
		}
		rows := make([]entities.Settlement, 0, len(changed))
		for k := range changed {
			if s, ok := global[k]; ok {
				rows = append(rows, *s)
			}
		}
		if len(rows) > 0 {
			if err := m.settlementRepo.UpsertBatch(jobCtx, rows, jobID); err != nil {
				return err
			}
			// Stream rows to CSV
			for _, s := range rows {
				_ = w.Write([]string{
					s.MerchantID,
					s.Date.Format("2006-01-02"),
					strconv.FormatInt(s.GrossCents, 10),
					strconv.FormatInt(s.FeeCents, 10),
					strconv.FormatInt(s.NetCents, 10),
					strconv.FormatInt(s.TxnCount, 10),
				})
			}
			w.Flush()
			if err := w.Error(); err != nil {
				return err
			}
		}
		// Update progress after each flush
		progress := 0
		if total > 0 {
			progress = int((processed * 100) / total)
			if progress > 100 {
				progress = 100
			}
		} else {
			progress = 100
		}
		_ = m.jobRepo.UpdateProgress(jobCtx, jobID, processed, total, progress)
		// Reset trackers
		changed = make(map[string]struct{})
		batchesSinceFlush = 0
		return nil
	}

	// Main collect loop
	for {
		select {
		case <-jobCtx.Done():
			// Try to mark cancelled and return
			_ = m.jobRepo.SetStatus(context.Background(), jobID, jobStatusCancelled)
			_ = m.jobRepo.SetResultPath(context.Background(), jobID, outPath)
			return
		case err := <-producerErr:
			if err != nil {
				m.fail(jobCtx, jobID, fmt.Errorf("producer: %w", err))
				return
			}
		default:
		}

		pr, ok := <-resultChan
		if !ok {
			// final flush
			if err := flush(true); err != nil {
				m.fail(jobCtx, jobID, fmt.Errorf("final flush: %w", err))
				return
			}
			// success
			_ = m.jobRepo.UpdateProgress(jobCtx, jobID, total, total, 100)
			_ = m.jobRepo.SetResultPath(jobCtx, jobID, outPath)
			_ = m.jobRepo.SetStatus(jobCtx, jobID, jobStatusCompleted)
			return
		}

		// Merge partial
		for k, v := range pr.agg {
			if cur, ok := global[k]; ok {
				cur.GrossCents += v.GrossCents
				cur.FeeCents += v.FeeCents
				cur.NetCents += v.NetCents
				cur.TxnCount += v.TxnCount
				global[k] = cur
			} else {
				vv := v // create local copy
				global[k] = &vv
			}
			changed[k] = struct{}{}
		}
		processed += int64(pr.count)
		batchesSinceFlush++
		if batchesSinceFlush >= flushEveryBatches {
			if err := flush(false); err != nil {
				m.fail(jobCtx, jobID, fmt.Errorf("flush: %w", err))
				return
			}
		}
	}
}

func (m *JobManager) fail(ctx context.Context, jobID string, err error) {
	_ = m.jobRepo.SetStatus(ctx, jobID, jobStatusFailed)
	// Best-effort progress update remains whatever it was.
	// Result path may still be set if CSV was created.
	_ = m.jobRepo.SetResultPath(ctx, jobID, filepath.Join("/tmp/settlements", jobID+".csv"))
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
