package repository

import (
    "context"

    "github.com/Caknoooo/go-gin-clean-starter/database/entities"
    "gorm.io/gorm"
)

type JobRepo interface {
    Create(ctx context.Context, job entities.Job) error
    UpdateProgress(ctx context.Context, jobID string, processed, total int64, progress int) error
    SetResultPath(ctx context.Context, jobID, path string) error
    RequestCancel(ctx context.Context, jobID string) error
    IsCancelRequested(ctx context.Context, jobID string) (bool, error)
    Get(ctx context.Context, jobID string) (entities.Job, error)
}

type jobRepository struct {
    db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepo {
    return &jobRepository{db: db}
}

func (r *jobRepository) Create(ctx context.Context, job entities.Job) error {
    return r.db.WithContext(ctx).Create(&job).Error
}

func (r *jobRepository) UpdateProgress(ctx context.Context, jobID string, processed, total int64, progress int) error {
    return r.db.WithContext(ctx).Model(&entities.Job{}).
        Where("id = ?", jobID).
        Updates(map[string]interface{}{
            "processed": processed,
            "total":     total,
            "progress":  progress,
        }).Error
}

func (r *jobRepository) SetResultPath(ctx context.Context, jobID, path string) error {
    return r.db.WithContext(ctx).Model(&entities.Job{}).
        Where("id = ?", jobID).
        Update("result_path", path).Error
}

func (r *jobRepository) RequestCancel(ctx context.Context, jobID string) error {
    return r.db.WithContext(ctx).Model(&entities.Job{}).
        Where("id = ?", jobID).
        Update("cancel_requested", true).Error
}

func (r *jobRepository) IsCancelRequested(ctx context.Context, jobID string) (bool, error) {
    var j entities.Job
    if err := r.db.WithContext(ctx).Model(&entities.Job{}).
        Select("cancel_requested").
        Where("id = ?", jobID).
        Take(&j).Error; err != nil {
        return false, err
    }
    return j.CancelRequested, nil
}

func (r *jobRepository) Get(ctx context.Context, jobID string) (entities.Job, error) {
    var j entities.Job
    if err := r.db.WithContext(ctx).Where("id = ?", jobID).Take(&j).Error; err != nil {
        return entities.Job{}, err
    }
    return j, nil
}
