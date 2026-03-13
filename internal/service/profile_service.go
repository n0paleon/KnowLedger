package service

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/repository"
	"KnowLedger/pkg/utils"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type ProfileService struct {
	db              *gorm.DB
	adminRepository *repository.AdminRepository
}

func NewProfileService(db *gorm.DB, adminRepository *repository.AdminRepository) *ProfileService {
	return &ProfileService{
		db:              db,
		adminRepository: adminRepository,
	}
}

func (s *ProfileService) GetProfileDetails(ctx context.Context, userID string) (*model.Admin, error) {
	profile, err := s.adminRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile details: %w", err)
	}
	return profile, nil
}

func (s *ProfileService) ResetApiKey(ctx context.Context, userID string) (string, error) {
	newApiKey, err := utils.GenerateAPIKey(32, "knowledger_j")
	if err != nil {
		return "", fmt.Errorf("failed to generate api key: %w", err)
	}

	if err := s.adminRepository.UpdateApiKeyByUserID(ctx, userID, newApiKey); err != nil {
		return "", fmt.Errorf("failed to update api key: %w", err)
	}

	return newApiKey, nil
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID, password, confirmPassword string) error {
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	hashedPassword, err := utils.GeneratePasswordHash(password)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	if err := s.adminRepository.UpdatePasswordByUserID(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
