package service

import (
	"KnowLedger/internal/repository"

	"gorm.io/gorm"
)

type InternalApiService struct {
	db             *gorm.DB
	factRepository *repository.FactRepository
	tagRepository  *repository.TagRepository
}

func NewInternalApiService(db *gorm.DB, factRepo *repository.FactRepository, tagRepository *repository.TagRepository) *InternalApiService {
	return &InternalApiService{
		db:             db,
		factRepository: factRepo,
		tagRepository:  tagRepository,
	}
}
