package repository

import (
	"KnowLedger/internal/model"
	"context"

	"gorm.io/gorm"
)

type AdminRepository struct {
	Repository[AdminRepository]
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{
		Repository[AdminRepository]{
			db: db,
			factory: func(tx *gorm.DB) *AdminRepository {
				return &AdminRepository{
					Repository[AdminRepository]{
						db: tx,
					},
				}
			},
		},
	}
}

func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.WithContext(ctx).
		Where("username = ?", username).First(&admin).Error
	return &admin, err
}

func (r *AdminRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Admin{}).Count(&count).Error
	return count, err
}

func (r *AdminRepository) Create(ctx context.Context, username, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Create(&model.Admin{
			Username: username,
			Password: hashedPassword,
		}).Error
}
