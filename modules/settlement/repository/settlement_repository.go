package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
)

type SettlementRepo interface {
	UpsertBatch(ctx context.Context, settlements []entities.Settlement, runID string) error
}
