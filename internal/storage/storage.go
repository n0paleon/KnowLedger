package storage

import (
	"KnowLedger/internal/model"
	"context"
)

type Storage interface {
	Upload(ctx context.Context, data []byte) (*model.MediaItem, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetURL(ctx context.Context, key string) string
	GetDetails(ctx context.Context, key string) (*model.MediaItem, error)
	DeleteBatch(ctx context.Context, keys []string) error
	ScanAll(ctx context.Context, fn func(item ScanResult) error) error
}

type ScanResult struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
}
