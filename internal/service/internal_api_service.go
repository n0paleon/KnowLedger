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
	"sync"
	"sync/atomic"

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

func (s *InternalApiService) CreateFunFact(ctx context.Context, facts []*dto.CreateFunFactAPIRequest) (int, error) {
	errCh := make(chan error, len(facts))
	var wg sync.WaitGroup
	var successCount atomic.Int32

	for _, fact := range facts {
		fact := fact
		wg.Add(1)
		_ = s.pool.Submit(func() {
			defer wg.Done()

			tags := strings.Join(fact.Tags, ",")
			req := &dto.PostCreateFunFactRequest{
				Content:   fact.Content,
				Tags:      tags,
				Status:    model.FactStatusDraft,
				SourceURL: fact.SourceURL,
				MediaKey:  fact.MediaKey,
			}
			if err := s.funFactService.CreateFact(ctx, req); err != nil {
				errCh <- fmt.Errorf("failed to create fact (content: %.30q): %w", fact.Content, err)
				return
			}
			successCount.Add(1)
		})
	}

	wg.Wait()
	close(errCh)

	var multiErr error
	for err := range errCh {
		multiErr = errors.Join(multiErr, err)
	}

	return int(successCount.Load()), multiErr
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
