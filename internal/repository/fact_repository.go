package repository

import (
	"KnowLedger/internal/model"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FactRepository struct {
	Repository[FactRepository]
}

func NewFactRepository(db *gorm.DB) *FactRepository {
	return &FactRepository{
		Repository[FactRepository]{
			db: db,
			factory: func(tx *gorm.DB) *FactRepository {
				return &FactRepository{
					Repository[FactRepository]{
						db: tx,
					},
				}
			},
		},
	}
}

// --- Scopes ---

func factWithSearch(search string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if search == "" {
			return db
		}
		return db.Where("content ILIKE ?", "%"+search+"%")
	}
}

func factWithStatus(status model.FactStatus) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == "" {
			return db
		}
		return db.Where("status = ?", status)
	}
}

func factWithSort(sortBy, sortDir string) func(db *gorm.DB) *gorm.DB {
	allowed := map[string]bool{
		"created_at": true,
		"updated_at": true,
	}
	return func(db *gorm.DB) *gorm.DB {
		if !allowed[sortBy] {
			sortBy = "created_at"
		}
		if sortDir != "asc" && sortDir != "desc" {
			sortDir = "desc"
		}
		return db.Order(sortBy + " " + sortDir)
	}
}

func factWithPagination(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 || limit > 100 {
			limit = 10
		}
		return db.Limit(limit).Offset((page - 1) * limit)
	}
}

// --- Repository Methods

func (r *FactRepository) Create(ctx context.Context, fact *model.Fact) error {
	return r.db.WithContext(ctx).Create(fact).Error
}

func (r *FactRepository) FindByID(ctx context.Context, id string) (*model.Fact, error) {
	var fact model.Fact

	err := r.db.WithContext(ctx).Preload("Tags").First(&fact, "id = ?", id).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find fact by id: %w", err)
	}

	return &fact, nil
}

func (r *FactRepository) Count(ctx context.Context) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).Model(&model.Fact{}).Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count fact: %w", err)
	}

	return total, nil
}

func (r *FactRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Fact{}).Error
}

func (r *FactRepository) GetRandomOne(ctx context.Context) (*model.Fact, error) {
	var fact model.Fact

	err := r.db.WithContext(ctx).Preload("Tags").Where("status = ?", model.FactStatusPublished).Order("RANDOM()").First(&fact).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find one random funfact: %w", err)
	}

	return &fact, nil
}

func (r *FactRepository) GetFacts(ctx context.Context, params model.ListFactsParams) (*model.Paginated[*model.Fact], error) {
	var facts []*model.Fact
	var total int64

	base := r.db.WithContext(ctx).Model(&model.Fact{}).
		Scopes(
			factWithSearch(params.Search),
			factWithStatus(params.Status),
		)

	if err := base.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count facts: %w", err)
	}

	err := base.
		Scopes(
			factWithSort(params.SortBy, params.SortDir),
			factWithPagination(params.Page, params.Limit),
		).
		Preload("Tags").
		Find(&facts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find facts: %w", err)
	}

	return model.NewPaginated(facts, total, params.Page, params.Limit), nil
}

// NOTE: Sudah dioptimasi untuk menghindari N+1 query
func (r *FactRepository) Update(
	ctx context.Context,
	id string,
	updatedFact *model.Fact,
	tagNames []string,
) (*model.Fact, error) {
	var result model.Fact

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Fact{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check fact: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("fact not found")
		}

		updates := map[string]interface{}{
			"content":    updatedFact.Content,
			"status":     updatedFact.Status,
			"source_url": updatedFact.SourceURL,
			"media":      updatedFact.Media,
		}
		if err := tx.Model(&model.Fact{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update fact: %w", err)
		}

		if err := tx.Where("fact_id = ?", id).Delete(&model.FactTag{}).Error; err != nil {
			return fmt.Errorf("failed to clear old tags: %w", err)
		}

		if len(tagNames) > 0 {
			var existingTags []model.Tag
			if err := tx.Where("name IN ?", tagNames).Find(&existingTags).Error; err != nil {
				return fmt.Errorf("failed to fetch existing tags: %w", err)
			}

			existingMap := make(map[string]model.Tag, len(existingTags))
			for _, t := range existingTags {
				existingMap[t.Name] = t
			}

			var newTags []model.Tag
			for _, name := range tagNames {
				if _, found := existingMap[name]; !found {
					newTags = append(newTags, model.Tag{Name: name})
				}
			}

			if len(newTags) > 0 {
				if err := tx.
					Clauses(clause.OnConflict{DoNothing: true}).
					Create(&newTags).Error; err != nil {
					return fmt.Errorf("failed to create new tags: %w", err)
				}
			}

			var allTags []model.Tag
			if err := tx.Where("name IN ?", tagNames).Find(&allTags).Error; err != nil {
				return fmt.Errorf("failed to fetch all tags: %w", err)
			}

			factTags := make([]model.FactTag, 0, len(allTags))
			for _, tag := range allTags {
				factTags = append(factTags, model.FactTag{
					FactID: id,
					TagID:  tag.ID,
				})
			}
			if err := tx.
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&factTags).Error; err != nil {
				return fmt.Errorf("failed to insert fact_tags: %w", err)
			}
		}

		return tx.Preload("Tags").First(&result, "id = ?", id).Error
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *FactRepository) ScanMediaKeys(ctx context.Context, onKey func(key string) error) error {
	var batch []model.Fact

	cutoff := time.Now().Add(-1 * time.Hour)

	result := r.db.WithContext(ctx).
		Model(&model.Fact{}).
		Where("media IS NOT NULL AND created_at <= ? AND updated_at <= ?", cutoff, cutoff).
		FindInBatches(&batch, 100, func(tx *gorm.DB, batchNum int) error {
			for _, f := range batch {
				if f.Media != nil {
					if err := onKey(f.Media.Key); err != nil {
						return err
					}
				}
			}
			return nil
		})

	return result.Error
}
