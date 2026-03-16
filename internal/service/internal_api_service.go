package service

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/repository"
	"KnowLedger/internal/workerpool"
	"KnowLedger/pkg/dto"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type InternalApiService struct {
	db             *gorm.DB
	factRepository *repository.FactRepository
	tagRepository  *repository.TagRepository
	mediaService   *MediaService
	funFactService *FunFactService
	profileService *ProfileService
	pool           *workerpool.Pool
}

func NewInternalApiService(
	db *gorm.DB,
	mediaService *MediaService,
	funFactService *FunFactService,
	profileService *ProfileService,
	pool *workerpool.Pool,
) *InternalApiService {
	return &InternalApiService{
		db:             db,
		mediaService:   mediaService,
		funFactService: funFactService,
		profileService: profileService,
		pool:           pool,
	}
}

func (s *InternalApiService) SaveMedia(ctx context.Context, data []byte) (*model.MediaItem, error) {
	return s.mediaService.SaveMedia(ctx, data)
}

func (s *InternalApiService) CreateFunFacts(ctx context.Context, facts []*dto.CreateFunFactAPIRequest) (failed int, errors error) {
	reqs := make([]*dto.PostCreateFunFactRequest, 0, len(facts))

	for _, f := range facts {
		// TODO: add data filters here to minimize failed queries that are not tracked due to query batching
		reqs = append(reqs, &dto.PostCreateFunFactRequest{
			Content:   f.Content,
			Tags:      strings.Join(f.Tags, ","),
			Status:    model.FactStatusDraft,
			SourceURL: f.SourceURL,
			MediaKey:  f.MediaKey,
		})
	}
	return s.funFactService.CreateFactBulk(ctx, reqs)
}

func (s *InternalApiService) ValidateApiKey(ctx context.Context, userID, apiKey string) error {
	admin, err := s.profileService.GetProfileDetails(ctx, userID)
	if err != nil {
		return fmt.Errorf("invalid api key: %w", err)
	}
	if admin.ApiKey != apiKey {
		return errors.New("invalid api key")
	}
	return nil
}
