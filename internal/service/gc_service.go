package service

import (
	"KnowLedger/internal/repository"
	"KnowLedger/internal/storage"
	"KnowLedger/internal/workerpool"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// TODO: garbage collector service
// TODO: add more cleanup methods (database, storage, cache, etc)

type GCService struct {
	factRepository *repository.FactRepository
	storage        storage.FileStorage
	pool           *workerpool.Pool
	log            *zap.Logger
	interval       time.Duration
	stopCh         chan struct{}
}

type GCServiceConfig struct {
	FactRepository *repository.FactRepository
	Storage        storage.FileStorage
	Pool           *workerpool.Pool
	Log            *zap.Logger
	Interval       time.Duration
}

func NewGCService(config GCServiceConfig) *GCService {
	interval := config.Interval
	if interval == 0 {
		interval = 1 * time.Hour
	}

	return &GCService{
		factRepository: config.FactRepository,
		storage:        config.Storage,
		pool:           config.Pool,
		log:            config.Log,
		interval:       interval,
		stopCh:         make(chan struct{}),
	}
}

func (s *GCService) Start() {
	go s.run()
}

func (s *GCService) Stop() {
	close(s.stopCh)
}

func (s *GCService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("gc service started", zap.String("interval", s.interval.String()))

	for {
		select {
		case <-ticker.C:
			s.cleanupObjectStorage()
		case <-s.stopCh:
			s.log.Info("gc service stopped", zap.String("interval", s.interval.String()))
			return
		}
	}
}

// cleanup scan all object on storage, filter unused object, and delete them
func (s *GCService) cleanupObjectStorage() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // max 5 minute for waiting cleanup processes
	defer cancel()

	s.log.Info("gc cleanup started")

	keySet := make(map[string]struct{})

	err := s.factRepository.ScanMediaKeys(ctx, func(key string) error {
		keySet[key] = struct{}{}
		return nil
	})

	if err != nil {
		s.log.Error("failed to scan media keys", zap.Error(err))
		return
	}

	var orphanKeys []string
	if err := s.storage.ScanAll(ctx, func(item storage.ScanResult) error {
		if time.Since(item.LastModified) < time.Hour {
			return nil
		}

		if _, known := keySet[item.Key]; !known {
			orphanKeys = append(orphanKeys, item.Key)
		}
		return nil
	}); err != nil {
		s.log.Error("failed to scan orphan keys", zap.Error(err))
		return
	}

	if len(orphanKeys) == 0 {
		s.log.Info("gc no orphan objects found")
		return
	}

	s.log.Info("gc found orphan objects", zap.Int("count", len(orphanKeys)))

	const batchSize = 1000
	var wg sync.WaitGroup

	for i := 0; i < len(orphanKeys); i += batchSize {
		end := min(i+batchSize, len(orphanKeys))
		batch := orphanKeys[i:end]

		wg.Add(1)
		_ = s.pool.Submit(func() {
			defer wg.Done()

			if err := s.storage.DeleteBatch(ctx, batch); err != nil {
				s.log.Error("gc batch delete failed",
					zap.Int("batch_start", i),
					zap.Int("batch_size", len(batch)),
					zap.Error(err),
				)
			} else {
				s.log.Info("gc batch delete finished",
					zap.Int("batch_size", len(batch)),
				)
			}
		})
	}

	wg.Wait()
}
