package repository

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
)

type JobRepo interface {
	Create(ctx context.Context, job entities.Job) error
	UpdateProgress(ctx context.Context, jobID string, processed, total int64, progress int) error
	SetResultPath(ctx context.Context, jobID, path string) error
	RequestCancel(ctx context.Context, jobID string) error
	IsCancelRequested(ctx context.Context, jobID string) (bool, error)
	Get(ctx context.Context, jobID string) (entities.Job, error)
}
