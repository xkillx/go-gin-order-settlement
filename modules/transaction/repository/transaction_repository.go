package repository

import (
	"context"
	"time"

	"github.com/xkillx/go-gin-order-settlement/database/entities"
	"gorm.io/gorm"
)

type TransactionRepo interface {
	Count(ctx context.Context, from, to time.Time) (int64, error)
	StreamByDateRange(ctx context.Context, from, to time.Time, batchSize int, out chan<- []entities.Transaction) error
}
type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepo {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Count(ctx context.Context, from, to time.Time) (int64, error) {
	var cnt int64
	err := r.db.WithContext(ctx).
		Model(&entities.Transaction{}).
		Where("paid_at >= ? AND paid_at < ?", from, to).
		Count(&cnt).Error
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *transactionRepository) StreamByDateRange(
	ctx context.Context,
	from, to time.Time,
	batchSize int,
	out chan<- []entities.Transaction,
) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	offset := 0
	for {
		// Respect context cancellation between batches
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var batch []entities.Transaction
		err := r.db.WithContext(ctx).
			Where("paid_at >= ? AND paid_at < ?", from, to).
			Order("paid_at ASC, id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&batch).Error
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}

		// Send the batch; caller owns channel lifecycle
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
