package repository

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettlementRepo interface {
	UpsertBatch(ctx context.Context, settlements []entities.Settlement, runID string) error
}

type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) SettlementRepo {
	return &settlementRepository{db: db}
}

// UpsertBatch inserts or updates settlements based on the unique (merchant_id, date) index.
// runID is accepted for traceability by callers but not persisted in the settlement rows.
func (r *settlementRepository) UpsertBatch(
	ctx context.Context,
	settlements []entities.Settlement,
	runID string,
) error {
	if len(settlements) == 0 {
		return nil
	}

	// Ensure UpdatedAt is set for DoUpdates AssignmentColumns to avoid zero timestamps
	now := time.Now().UTC()
	for i := range settlements {
		settlements[i].UpdatedAt = now
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "merchant_id"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"gross_cents", "fee_cents", "net_cents", "txn_count", "updated_at"}),
		}).
		Create(&settlements).Error
}
