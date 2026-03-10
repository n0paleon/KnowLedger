package service

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/storage"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type MediaService struct {
	storage storage.Storage
	log     *zap.Logger
}

func NewMediaService(storage storage.Storage, logger *zap.Logger) *MediaService {
	return &MediaService{
		storage: storage,
		log:     logger,
	}
}

func (s *MediaService) SaveMedia(ctx context.Context, data []byte) (*model.MediaItem, error) {
	start := time.Now()
	result, err := s.storage.Upload(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	s.logDebugPerformance(start, result)

	return result, nil
}

func (s *MediaService) GetMediaDetails(ctx context.Context, key string) (*model.MediaItem, error) {
	start := time.Now()
	result, err := s.storage.GetDetails(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get media details: %w", err)
	}

	s.logDebugPerformance(start, result)

	return result, nil
}

func (s *MediaService) logDebugPerformance(start time.Time, media *model.MediaItem) {
	if media == nil {
		return
	}
	s.log.Debug("media get details",
		zap.String("key", media.Key),
		zap.String("hash", media.Hash),
		zap.Int64("size", media.Size),
		zap.String("contentType", media.ContentType),
		zap.String("time_taken", time.Since(start).String()),
	)
}
