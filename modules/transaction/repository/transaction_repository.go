package repository

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
)

type TransactionRepo interface {
	Count(ctx context.Context, from, to time.Time) (int64, error)
	StreamByDateRange(ctx context.Context, from, to time.Time, batchSize int, out chan<- []entities.Transaction) error
}
