package settlement

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"

	"github.com/xkillx/go-gin-order-settlement/config"
	"github.com/xkillx/go-gin-order-settlement/database"
	"github.com/xkillx/go-gin-order-settlement/database/entities"
	seeds "github.com/xkillx/go-gin-order-settlement/database/seeders/seeds"
	jobrepo "github.com/xkillx/go-gin-order-settlement/modules/job/repository"
	settlement "github.com/xkillx/go-gin-order-settlement/modules/settlement"
	settrepo "github.com/xkillx/go-gin-order-settlement/modules/settlement/repository"
	settlementService "github.com/xkillx/go-gin-order-settlement/modules/settlement/service"
	txrepo "github.com/xkillx/go-gin-order-settlement/modules/transaction/repository"
	"github.com/xkillx/go-gin-order-settlement/pkg/constants"
)

// testEnv bundles common test wiring
type testEnv struct {
	server   *gin.Engine
	injector *do.Injector
	db       *gorm.DB
}

// truncateTables clears relevant tables before each test to ensure a clean state
func truncateTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM settlements").Error; err != nil {
		t.Fatalf("truncate settlements: %v", err)
	}
	if err := db.Exec("DELETE FROM jobs").Error; err != nil {
		t.Fatalf("truncate jobs: %v", err)
	}
	if err := db.Exec("DELETE FROM transactions").Error; err != nil {
		t.Fatalf("truncate transactions: %v", err)
	}
}

// ensureSeederEnv ensures seeds connect to the same DB as SetUpTestDatabaseConnection by setting env defaults
func ensureSeederEnv() {
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("DB_PASS") == "" {
		os.Setenv("DB_PASS", "password")
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "test_db")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_SSLMODE") == "" {
		os.Setenv("DB_SSLMODE", "disable")
	}
}

// seedWithSeeder uses the bulk seeder with small numbers to keep tests fast
func seedWithSeeder(t *testing.T) {
	t.Helper()
	ensureSeederEnv()
	// rows, merchants, days, batchSize
	if err := seeds.BulkTransactionSeeder(nil, 3000, 10, 3, 1000); err != nil {
		t.Fatalf("bulk seeder failed: %v", err)
	}
}

// slowTransactionRepository implements txrepo.TransactionRepo with a delay per batch
// so cancellation windows are deterministic in tests.
type slowTransactionRepository struct {
	db    *gorm.DB
	delay time.Duration
}

func (r *slowTransactionRepository) Count(ctx context.Context, from, to time.Time) (int64, error) {
	return txrepo.NewTransactionRepository(r.db).Count(ctx, from, to)
}

func (r *slowTransactionRepository) StreamByDateRange(ctx context.Context, from, to time.Time, batchSize int, out chan<- []entities.Transaction) error {
	if batchSize <= 0 {
		batchSize = 1000
	}
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var batch []entities.Transaction
		if err := r.db.WithContext(ctx).
			Where("paid_at >= ? AND paid_at < ?", from, to).
			Order("paid_at ASC, id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&batch).Error; err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}

		if r.delay > 0 {
			time.Sleep(r.delay)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- batch:
		}

		offset += len(batch)
		if len(batch) < batchSize {
			return nil
		}
	}
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Use local PostgreSQL test DB like order tests
	db := config.SetUpTestDatabaseConnection()
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// Build repos and JobManager (default tx repo)
	return newTestEnvWithTxRepo(t, db, txrepo.NewTransactionRepository(db))
}

func newTestEnvWithTxRepo(t *testing.T, db *gorm.DB, txRepo txrepo.TransactionRepo) testEnv {
	t.Helper()

	stRepo := settrepo.NewSettlementRepository(db)
	jobRepo := jobrepo.NewJobRepository(db)
	jobManager := settlementService.NewJobManager(txRepo, stRepo, jobRepo)

	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) { return db, nil })
	do.Provide(injector, func(i *do.Injector) (*settlementService.JobManager, error) { return jobManager, nil })

	server := gin.New()
	settlement.RegisterRoutes(server, injector)

	return testEnv{server: server, injector: injector, db: db}
}

func TestSettlementCreateJob(t *testing.T) {
	env := newTestEnv(t)
	truncateTables(t, env.db)
	seedWithSeeder(t)

	// Narrow date window to reduce processed rows (today only)
	today := time.Now().UTC().Format("2006-01-02")
	tomorrow := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	body := map[string]string{"from": today, "to": tomorrow}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/jobs/settlement", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"].(string) != "QUEUED" {
		t.Fatalf("expected status QUEUED, got %#v", resp["status"])
	}
	if s, ok := resp["job_id"].(string); !ok || s == "" {
		t.Fatalf("expected non-empty job_id, got %#v", resp["job_id"])
	}
}

func TestSettlementGetJob(t *testing.T) {
	env := newTestEnv(t)
	truncateTables(t, env.db)
	seedWithSeeder(t)

	// Create job over last 2 days
	fromDate := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
	toDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	b, _ := json.Marshal(map[string]string{"from": fromDate, "to": toDate})
	req := httptest.NewRequest(http.MethodPost, "/jobs/settlement", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("create expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var create map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &create)
	jobID := create["job_id"].(string)

	// Poll GET /jobs/:id until COMPLETED or timeout
	deadline := time.Now().Add(20 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		grec := httptest.NewRecorder()
		greq := httptest.NewRequest(http.MethodGet, "/jobs/"+jobID, nil)
		env.server.ServeHTTP(grec, greq)
		if grec.Code != http.StatusOK {
			t.Fatalf("get expected 200, got %d: %s", grec.Code, grec.Body.String())
		}
		_ = json.Unmarshal(grec.Body.Bytes(), &last)
		if last["status"].(string) == "COMPLETED" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if last["status"].(string) != "COMPLETED" {
		t.Fatalf("job did not complete, last: %#v", last)
	}
	if _, ok := last["download_url"].(string); !ok {
		t.Fatalf("expected download_url on completion, got: %#v", last)
	}
	// Ensure CSV file exists
	csvPath := filepath.Join("/tmp/settlements", jobID+".csv")
	if _, err := os.Stat(csvPath); err != nil {
		t.Fatalf("expected csv exists at %s, err=%v", csvPath, err)
	}
}

func TestSettlementCancelJob(t *testing.T) {
	// Small batch and single worker to make cancellation more likely
	prevBatch := os.Getenv("BATCH_SIZE")
	prevWorkers := os.Getenv("WORKERS")
	os.Setenv("BATCH_SIZE", "5")
	os.Setenv("WORKERS", "1")
	t.Cleanup(func() {
		os.Setenv("BATCH_SIZE", prevBatch)
		os.Setenv("WORKERS", prevWorkers)
	})

	// Use local PostgreSQL with slow streaming repo to ensure the job stays running long enough
	db := config.SetUpTestDatabaseConnection()
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove("test.db") })
	truncateTables(t, db)
	seedWithSeeder(t)
	env := newTestEnvWithTxRepo(t, db, &slowTransactionRepository{db: db, delay: 10 * time.Millisecond})

	// Create job
	fromDate := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
	toDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	b, _ := json.Marshal(map[string]string{"from": fromDate, "to": toDate})
	req := httptest.NewRequest(http.MethodPost, "/jobs/settlement", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	env.server.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("create expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var create map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &create)
	jobID := create["job_id"].(string)

	// Give the background worker a moment to transition to RUNNING and register cancel func
	time.Sleep(50 * time.Millisecond)

	// Request cancel
	crec := httptest.NewRecorder()
	creq := httptest.NewRequest(http.MethodPost, "/jobs/"+jobID+"/cancel", nil)
	env.server.ServeHTTP(crec, creq)
	if crec.Code != http.StatusAccepted {
		t.Fatalf("cancel expected 202, got %d: %s", crec.Code, crec.Body.String())
	}

	// Poll until CANCELLED
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		grec := httptest.NewRecorder()
		greq := httptest.NewRequest(http.MethodGet, "/jobs/"+jobID, nil)
		env.server.ServeHTTP(grec, greq)
		if grec.Code != http.StatusOK {
			t.Fatalf("get expected 200, got %d: %s", grec.Code, grec.Body.String())
		}
		var got map[string]any
		_ = json.Unmarshal(grec.Body.Bytes(), &got)
		st := got["status"].(string)
		if st == "CANCELLED" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("job did not reach CANCELLED in time")
}
