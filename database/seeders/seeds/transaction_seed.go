package seeds

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// ListTransactionSeeder runs a transactions seeder with safe defaults for local use.
func ListTransactionSeeder(db *gorm.DB) error {
	const (
		rows      = 1_000_000
		merchants = 200
		days      = 30
		batchSize = 10_000
	)
	return BulkTransactionSeeder(db, rows, merchants, days, batchSize)
}

// BulkTransactionSeeder seeds the public.transactions table efficiently using pgx COPY.
// It ignores the provided *gorm.DB for insertion (GORM is too slow for millions of rows),
// but keeping the signature makes it convenient to call from your existing seeder pipeline.
func BulkTransactionSeeder(_ *gorm.DB, rows, merchants, days, batchSize int) error {
	if rows <= 0 || merchants <= 0 || days <= 0 || batchSize <= 0 {
		return fmt.Errorf("invalid params: rows, merchants, days, batchSize must be > 0")
	}

	// Load .env if present
	_ = godotenv.Load(".env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := connectPool(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer pool.Close()

	if err := ensureTransactionsTableExists(ctx, pool); err != nil {
		return err
	}

	merchantIDs := make([]string, merchants)
	for i := 0; i < merchants; i++ {
		merchantIDs[i] = fmt.Sprintf("merchant-%d", i+1)
	}

	end := time.Now().UTC()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	cols := []string{"id", "merchant_id", "amount_cents", "fee_cents", "status", "paid_at", "created_at", "updated_at"}

	log.Printf("transaction seeder: rows=%d merchants=%d days=%d batch=%d", rows, merchants, days, batchSize)

	total := 0
	nextProgress := 50_000

	for total < rows {
		remaining := rows - total
		curBatch := batchSize
		if remaining < curBatch {
			curBatch = remaining
		}

		rowsBuf := make([][]any, 0, curBatch)
		for i := 0; i < curBatch; i++ {
			id := uuid.New()
			mid := merchantIDs[r.Intn(len(merchantIDs))]
			amount := int64(100 + r.Intn(200_000))
			fee := int64(math.Round(float64(amount)*0.029)) + 30
			paidAt := randomTimeBetween(r, start, end)
			createdAt := paidAt
			updatedAt := paidAt
			status := "paid"
			rowsBuf = append(rowsBuf, []any{id, mid, amount, fee, status, paidAt, createdAt, updatedAt})
		}

		inserted, err := pool.CopyFrom(ctx, pgx.Identifier{"public", "transactions"}, cols, pgx.CopyFromRows(rowsBuf))
		if err != nil {
			return fmt.Errorf("COPY failed at %d: %w", total, err)
		}
		total += int(inserted)

		if total >= nextProgress {
			log.Printf("transaction seeder progress: %d/%d (%.1f%%)", total, rows, 100*float64(total)/float64(rows))
			nextProgress += 50_000
		}
	}

	log.Printf("transaction seeder done: inserted=%d", total)
	return nil
}

func ensureTransactionsTableExists(ctx context.Context, pool *pgxpool.Pool) error {
	var reg *string
	if err := pool.QueryRow(ctx, "select to_regclass('public.transactions')").Scan(&reg); err != nil {
		return fmt.Errorf("check table: %w", err)
	}
	if reg == nil || *reg == "" {
		return fmt.Errorf("table public.transactions not found. Run migrations first")
	}
	// Best-effort column presence check
	cols := map[string]bool{}
	rows, err := pool.Query(ctx, `
        select column_name
        from information_schema.columns
        where table_schema='public' and table_name='transactions'
    `)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c string
			_ = rows.Scan(&c)
			cols[strings.ToLower(c)] = true
		}
		for _, need := range []string{"id", "merchant_id", "amount_cents", "fee_cents", "status", "paid_at", "created_at", "updated_at"} {
			if !cols[need] {
				return fmt.Errorf("missing required column '%s' on public.transactions", need)
			}
		}
	}
	return nil
}

func connectPool(ctx context.Context) (*pgxpool.Pool, error) {
	host := getenvDefault("DB_HOST", "localhost")
	user := getenvDefault("DB_USER", "postgres")
	pass := getenvDefault("DB_PASS", "")
	dbname := getenvDefault("DB_NAME", "postgres")
	port := getenvDefault("DB_PORT", "5432")
	sslmode := getenvDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, pass, dbname, port, sslmode)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = 4
	cfg.MaxConnLifetime = time.Minute * 10
	cfg.MaxConnIdleTime = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

func randomTimeBetween(r *rand.Rand, start, end time.Time) time.Time {
	if !end.After(start) {
		return start
	}
	delta := end.Sub(start)
	max := delta.Nanoseconds()
	if max <= 0 {
		return start
	}
	n := r.Int63n(max)
	return start.Add(time.Duration(n))
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
