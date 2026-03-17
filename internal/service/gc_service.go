package service

import (
	"KnowLedger/internal/repository"
	"KnowLedger/internal/storage"
	"KnowLedger/internal/workerpool"
	"KnowLedger/pkg/utils"
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type GCService struct {
	factRepository      *repository.FactRepository
	storage             storage.FileStorage
	pool                *workerpool.Pool
	log                 *zap.Logger
	interval            time.Duration
	stopCh              chan struct{}
	similarityThreshold float64
}

type GCServiceConfig struct {
	FactRepository      *repository.FactRepository
	Storage             storage.FileStorage
	Pool                *workerpool.Pool
	Log                 *zap.Logger
	Interval            time.Duration
	SimilarityThreshold float64
}

func NewGCService(config GCServiceConfig) *GCService {
	interval := config.Interval
	if interval == 0 {
		interval = 1 * time.Hour
	}

	return &GCService{
		factRepository:      config.FactRepository,
		storage:             config.Storage,
		pool:                config.Pool,
		log:                 config.Log,
		interval:            interval,
		stopCh:              make(chan struct{}),
		similarityThreshold: config.SimilarityThreshold,
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
			s.runAll()
		case <-s.stopCh:
			s.log.Info("gc service stopped", zap.String("interval", s.interval.String()))
			return
		}
	}
}

func (s *GCService) runAll() {
	s.cleanupObjectStorage()
	s.removeNearDuplicateFacts()
}

func (s *GCService) removeNearDuplicateFacts() {
	start := time.Now()
	timeout := 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.log.Info("trying to find duplicate/near-duplicate content", zap.Duration("timeout", timeout))

	facts, _ := s.factRepository.FindAll(ctx)
	if len(facts) == 0 {
		s.log.Info("no draft facts found")
		return
	}

	type normalizedFact struct {
		id      string
		text    string
		content string
	}
	normalized := make([]normalizedFact, len(facts))

	var wg sync.WaitGroup
	for i, fact := range facts {
		wg.Add(1)
		i, fact := i, fact
		_ = s.pool.Submit(func() {
			defer wg.Done()
			normalized[i] = normalizedFact{
				id:      fact.ID,
				text:    fact.NormalizedContent(),
				content: fact.Content,
			}
		})
	}
	wg.Wait()

	toDelete := make([]string, 0)
	type seen struct {
		id      string
		text    string
		content string
	}
	var seenFacts []seen

	for _, nf := range normalized {
		isDuplicate := false

		for i, sf := range seenFacts {
			similarityScore := utils.CosineSimilarityText(nf.text, sf.text)
			if similarityScore <= s.similarityThreshold {
				continue
			}

			if len(nf.content) >= len(sf.content) {
				toDelete = append(toDelete, sf.id)
				seenFacts[i] = seen{nf.id, nf.text, nf.content}
			} else {
				toDelete = append(toDelete, nf.id)
			}

			isDuplicate = true
			break
		}

		if !isDuplicate {
			seenFacts = append(seenFacts, seen{nf.id, nf.text, nf.content})
		}
	}

	if len(toDelete) > 0 {
		s.log.Info("removing duplicate facts",
			zap.Int("total", len(toDelete)),
			zap.Strings("ids", toDelete),
		)
		if err := s.factRepository.BulkDelete(ctx, toDelete); err != nil {
			s.log.Error("failed to delete facts", zap.Strings("factIds", toDelete), zap.Error(err))
		}
	} else {
		s.log.Info("no duplicate/near-duplicate facts found", zap.Int("total", len(toDelete)))
	}

	s.log.Debug("removeNearDuplicateFacts()",
		zap.Float64("time_taken_seconds", time.Since(start).Seconds()),
	)
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
