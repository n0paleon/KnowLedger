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

func (r *AdminRepository) FindAll(ctx context.Context) ([]model.Admin, error) {
	var admins []model.Admin

	if err := r.db.WithContext(ctx).Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
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

func (r *AdminRepository) Delete(ctx context.Context, admin model.Admin) error {
	return r.db.WithContext(ctx).Where(admin).Delete(&model.Admin{}).Error
}

func (r *AdminRepository) Create(ctx context.Context, username, hashedPassword string) (*model.Admin, error) {
	admin := model.Admin{
		Username: username,
		Password: hashedPassword,
	}

	err := r.db.WithContext(ctx).
		Create(&admin).Error

	return &admin, err
}

func (r *AdminRepository) UpdatePassword(ctx context.Context, username, newPassword string) error {
	return r.db.WithContext(ctx).
		Model(&model.Admin{}).
		Where("username = ?", username).
		Update("password", newPassword).Error
}
