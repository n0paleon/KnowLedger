package repository

import (
	"KnowLedger/internal/model"
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
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

func (r *TagRepository) GetTags(ctx context.Context, params model.ListTagsParams) (*model.Paginated[*model.Tag], error) {
	var (
		tags  []*model.Tag
		total int64
	)

	base := r.db.WithContext(ctx).Model(&model.Tag{})

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		countDB := base.Session(&gorm.Session{})
		if err := countDB.Count(&total).Error; err != nil {
			return fmt.Errorf("failed to count tags: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		dataDB := base.Session(&gorm.Session{})
		if err := dataDB.
			Select("tags.*, COUNT(fact_tags.fact_id) AS total_facts").
			Joins("LEFT JOIN fact_tags ON fact_tags.tag_id = tags.id").
			Group("tags.id").
			Order("tags.id ASC").
			Scopes(WithPagination(params.Page, params.Limit)).
			Find(&tags).Error; err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return model.NewPaginated(tags, total, params.Page, params.Limit), nil
}

func (r *TagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{}).Error
}

func (r *TagRepository) SearchTagByName(ctx context.Context, name string, limit int) ([]model.Tag, error) {
	var tags []model.Tag

	if err := r.db.WithContext(ctx).
		Scopes(WithSearch("name", name)).
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to search tag by name %w", err)
	}

	return tags, nil
}
