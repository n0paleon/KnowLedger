package repository

import (
	"KnowLedger/internal/model"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type TagRepository struct {
	Repository[TagRepository]
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{
		Repository[TagRepository]{
			db: db,
			factory: func(tx *gorm.DB) *TagRepository {
				return &TagRepository{
					Repository[TagRepository]{
						db: tx,
					},
				}
			},
		},
	}
}

func (r *TagRepository) FindTagsByNames(ctx context.Context, tagNames []string) ([]*model.Tag, error) {
	var tags []*model.Tag

	if err := r.db.WithContext(ctx).Where("name IN ?", tagNames).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get tags by names: %w", err)
	}

	return tags, nil
}

func (r *TagRepository) CreateBulk(ctx context.Context, tags *[]model.Tag) error {
	result := r.db.WithContext(ctx).CreateInBatches(tags, 20)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk create tags: %w", result.Error)
	}

	return nil
}

func (r *TagRepository) Count(ctx context.Context) (int64, error) {
	var total int64

	err := r.db.WithContext(ctx).Model(&model.Tag{}).Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count tag: %w", err)
	}

	return total, nil
}
