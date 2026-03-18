package repository

import (
	"KnowLedger/internal/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type GCJobRepository struct {
	Repository[GCJobRepository]
}

func NewGCJobRepository(db *gorm.DB) *GCJobRepository {
	return &GCJobRepository{
		Repository[GCJobRepository]{
			db: db,
			factory: func(tx *gorm.DB) *GCJobRepository {
				return &GCJobRepository{
					Repository[GCJobRepository]{
						db: tx,
					},
				}
			},
		},
	}
}

func gcJobWithStatus(status model.GCJobStatus) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == "" {
			return db
		}
		return db.Where("status = ?", status)
	}
}

func gcJobWithTrigger(trigger model.GCJobTrigger) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if trigger == "" {
			return db
		}
		return db.Where("trigger = ?", trigger)
	}
}

func (r *GCJobRepository) Create(ctx context.Context, job *model.GCJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *GCJobRepository) UpdateStatus(ctx context.Context, id string, status model.GCJobStatus, finishedAt *time.Time) error {
	updates := map[string]any{"status": status}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	return r.db.WithContext(ctx).Model(&model.GCJob{}).Where("id = ?", id).Updates(updates).Error
}

func (r *GCJobRepository) AppendLog(ctx context.Context, jobID, level, message string) error {
	return r.db.WithContext(ctx).Create(&model.GCJobLog{
		JobID:   jobID,
		Level:   level,
		Message: message,
	}).Error
}

func (r *GCJobRepository) FindAll(ctx context.Context, params model.GCJobListParams) (*model.Paginated[*model.GCJob], error) {
	var (
		jobs  []*model.GCJob
		total int64
	)

	base := r.db.WithContext(ctx).Model(&model.GCJob{}).Scopes(
		gcJobWithStatus(params.Status),
		gcJobWithTrigger(params.Trigger),
	)

	if err := base.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count jobs: %w", err)
	}

	err := base.
		Scopes(
			WithPagination(params.Page, params.Limit),
		).
		Order("created_at " + params.SortDir).
		Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}

	return model.NewPaginated(jobs, total, params.Page, params.Limit), nil
}

func (r *GCJobRepository) FindByID(ctx context.Context, id string) (*model.GCJob, error) {
	var job model.GCJob
	err := r.db.WithContext(ctx).Preload("Logs", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).First(&job, "id = ?", id).Error
	return &job, err
}

func (r *GCJobRepository) FindOlderThan(ctx context.Context, before time.Time) ([]*model.GCJob, error) {
	var jobs []*model.GCJob
	err := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Find(&jobs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch old jobs: %w", err)
	}
	return jobs, nil
}

func (r *GCJobRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.GCJob{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete jobs: %w", err)
	}
	return nil
}
