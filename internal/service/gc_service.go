package service

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/repository"
	"KnowLedger/internal/storage"
	"KnowLedger/internal/workerpool"
	"KnowLedger/pkg/dto"
	"KnowLedger/pkg/utils"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type GCService struct {
	factRepository      *repository.FactRepository
	gcJobRepository     *repository.GCJobRepository
	storage             storage.FileStorage
	pool                *workerpool.Pool
	log                 *zap.Logger
	interval            time.Duration
	stopCh              chan struct{}
	similarityThreshold float64
	mu                  sync.Mutex
	running             bool
	logRetention        time.Duration
	jobWg               sync.WaitGroup // tracking job yang sedang berjalan
	emitWg              sync.WaitGroup // tracking pending log writes ke DB
}

type GCServiceConfig struct {
	FactRepository      *repository.FactRepository
	GCJobRepository     *repository.GCJobRepository
	Storage             storage.FileStorage
	Pool                *workerpool.Pool
	Log                 *zap.Logger
	Interval            time.Duration
	SimilarityThreshold float64
	LogRetention        time.Duration
}

func NewGCService(config GCServiceConfig) *GCService {
	interval := config.Interval
	if interval == 0 {
		interval = 1 * time.Hour
	}

	return &GCService{
		factRepository:      config.FactRepository,
		gcJobRepository:     config.GCJobRepository,
		storage:             config.Storage,
		pool:                config.Pool,
		log:                 config.Log,
		interval:            interval,
		stopCh:              make(chan struct{}),
		similarityThreshold: config.SimilarityThreshold,
		logRetention:        config.LogRetention,
	}
}

func (s *GCService) Start() {
	go s.run()
}

func (s *GCService) Stop() {
	s.log.Info("gc service shutting down, waiting for active job to finish...")
	close(s.stopCh)

	done := make(chan struct{})
	go func() {
		s.jobWg.Wait()  // tunggu job selesai dulu
		s.emitWg.Wait() // baru tunggu semua log selesai ditulis ke DB
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("gc service shutdown complete")
	case <-time.After(10 * time.Minute):
		s.log.Warn("gc service shutdown timed out, forcing exit")
	}
}

func (s *GCService) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("gc service started", zap.String("interval", s.interval.String()))

	for {
		select {
		case <-s.stopCh:
			s.log.Info("gc service stopped")
			return
		case <-ticker.C:
			// Cek ulang stopCh sebelum trigger — hindari race saat keduanya ready bersamaan
			select {
			case <-s.stopCh:
				s.log.Info("gc service stopped")
				return
			default:
				s.triggerJob(context.Background(), model.GCJobTriggerAutomatic)
			}
		}
	}
}

// TriggerManual dipanggil oleh HTTP handler — non-blocking, return jobID langsung
func (s *GCService) TriggerManual(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return "", errors.New("gc is already running")
	}
	s.running = true
	s.mu.Unlock()

	job, err := s.createJob(ctx, model.GCJobTriggerManual)
	if err != nil {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return "", fmt.Errorf("failed to create gc job: %w", err)
	}

	s.jobWg.Add(1)
	_ = s.pool.Submit(func() {
		defer s.jobWg.Done()
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}()
		s.executeJob(context.Background(), job.ID)
	})

	return job.ID, nil
}

func (s *GCService) GetJobs(ctx context.Context, params *dto.GCJobListParams) (*model.Paginated[*model.GCJob], error) {
	if params.SortDir == "" {
		params.SortDir = "desc"
	}

	jobs, err := s.gcJobRepository.FindAll(ctx, model.GCJobListParams{
		Page:    params.Page,
		Limit:   params.Limit,
		Status:  params.Status,
		Trigger: params.Trigger,
		SortDir: params.SortDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	return jobs, nil
}

func (s *GCService) GetJobDetails(ctx context.Context, id string) (*model.GCJob, error) {
	return s.gcJobRepository.FindByID(ctx, id)
}

// triggerJob dipanggil oleh cron ticker — blocking sampai selesai
func (s *GCService) triggerJob(ctx context.Context, trigger model.GCJobTrigger) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.log.Info("gc skipped: already running")
		return
	}
	s.running = true
	s.mu.Unlock()

	s.jobWg.Add(1)
	defer s.jobWg.Done()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	job, err := s.createJob(ctx, trigger)
	if err != nil {
		s.log.Error("failed to create gc job", zap.Error(err))
		return
	}

	s.executeJob(ctx, job.ID)
}

func (s *GCService) createJob(ctx context.Context, trigger model.GCJobTrigger) (*model.GCJob, error) {
	now := time.Now()
	job := &model.GCJob{
		ID:        utils.GenerateRandomULID(),
		Status:    model.GCJobStatusRunning,
		Trigger:   trigger,
		StartedAt: &now,
	}
	return job, s.gcJobRepository.Create(ctx, job)
}

func (s *GCService) finishJob(ctx context.Context, jobID string, status model.GCJobStatus) {
	now := time.Now()
	if err := s.gcJobRepository.UpdateStatus(ctx, jobID, status, &now); err != nil {
		s.log.Error("failed to update gc job status",
			zap.String("job_id", jobID),
			zap.Error(err),
		)
	}
}

// emit menulis log ke zap dan ke DB secara fire-and-forget
func (s *GCService) emit(ctx context.Context, jobID, level, msg string) {
	switch level {
	case "error":
		s.log.Error(msg, zap.String("job_id", jobID))
	case "debug":
		s.log.Debug(msg, zap.String("job_id", jobID))
	default:
		s.log.Info(msg, zap.String("job_id", jobID))
	}

	s.emitWg.Add(1)
	_ = s.pool.Submit(func() {
		defer s.emitWg.Done()
		if err := s.gcJobRepository.AppendLog(context.Background(), jobID, level, msg); err != nil {
			s.log.Error("failed to append gc job log",
				zap.String("job_id", jobID),
				zap.Error(err),
			)
		}
	})
}

func (s *GCService) executeJob(ctx context.Context, jobID string) {
	status := model.GCJobStatusCompleted
	defer func() {
		s.finishJob(ctx, jobID, status)
	}()

	if err := s.cleanupObjectStorage(ctx, jobID); err != nil {
		s.emit(ctx, jobID, "error", fmt.Sprintf("object storage cleanup failed: %v", err))
		status = model.GCJobStatusFailed
		return
	}

	s.removeNearDuplicateFacts(ctx, jobID)
	s.cleanupOldJobs(ctx, jobID)
}

func (s *GCService) cleanupObjectStorage(ctx context.Context, jobID string) error {
	cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	s.emit(cleanupCtx, jobID, "info", "scanning media keys from database...")

	keySet := make(map[string]struct{})
	if err := s.factRepository.ScanMediaKeys(cleanupCtx, func(key string) error {
		keySet[key] = struct{}{}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to scan media keys: %w", err)
	}

	s.emit(cleanupCtx, jobID, "info", fmt.Sprintf("found %d known media keys, scanning object storage...", len(keySet)))

	var orphanKeys []string
	if err := s.storage.ScanAll(cleanupCtx, func(item storage.ScanResult) error {
		if time.Since(item.LastModified) < time.Hour {
			return nil
		}
		if _, known := keySet[item.Key]; !known {
			orphanKeys = append(orphanKeys, item.Key)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to scan object storage: %w", err)
	}

	if len(orphanKeys) == 0 {
		s.emit(cleanupCtx, jobID, "info", "no orphan objects found")
		return nil
	}

	s.emit(cleanupCtx, jobID, "info", fmt.Sprintf("found %d orphan objects, deleting in batches...", len(orphanKeys)))

	const batchSize = 1000
	var wg sync.WaitGroup
	var deletedCount sync.Map

	for i := 0; i < len(orphanKeys); i += batchSize {
		end := min(i+batchSize, len(orphanKeys))
		batch := orphanKeys[i:end]
		batchStart := i

		wg.Add(1)
		_ = s.pool.Submit(func() {
			defer wg.Done()

			if err := s.storage.DeleteBatch(cleanupCtx, batch); err != nil {
				s.emit(cleanupCtx, jobID, "error", fmt.Sprintf(
					"batch delete failed (start: %d, size: %d): %v",
					batchStart, len(batch), err,
				))
			} else {
				deletedCount.Store(batchStart, len(batch))
				s.emit(cleanupCtx, jobID, "info", fmt.Sprintf(
					"batch deleted (start: %d, size: %d)",
					batchStart, len(batch),
				))
			}
		})
	}

	wg.Wait()

	s.emit(cleanupCtx, jobID, "info", "object storage cleanup completed")
	return nil
}

func (s *GCService) removeNearDuplicateFacts(ctx context.Context, jobID string) {
	start := time.Now()
	dupCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	s.emit(dupCtx, jobID, "info", "loading facts for duplicate detection...")

	facts, err := s.factRepository.FindAll(dupCtx)
	if err != nil {
		s.emit(dupCtx, jobID, "error", fmt.Sprintf("failed to load facts: %v", err))
		return
	}

	if len(facts) == 0 {
		s.emit(dupCtx, jobID, "info", "no facts found, skipping duplicate detection")
		return
	}

	s.emit(dupCtx, jobID, "info", fmt.Sprintf("loaded %d facts, normalizing content in parallel...", len(facts)))

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

	s.emit(dupCtx, jobID, "info", "normalization done, running similarity check...")

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
			score := utils.CosineSimilarityText(nf.text, sf.text)
			if score <= s.similarityThreshold {
				continue
			}

			if len(nf.content) >= len(sf.content) {
				s.emit(dupCtx, jobID, "info", fmt.Sprintf(
					"near-duplicate: keep %s, remove %s (score: %.4f)",
					nf.id, sf.id, score,
				))
				toDelete = append(toDelete, sf.id)
				seenFacts[i] = seen{nf.id, nf.text, nf.content}
			} else {
				s.emit(dupCtx, jobID, "info", fmt.Sprintf(
					"near-duplicate: keep %s, remove %s (score: %.4f)",
					sf.id, nf.id, score,
				))
				toDelete = append(toDelete, nf.id)
			}

			isDuplicate = true
			break
		}

		if !isDuplicate {
			seenFacts = append(seenFacts, seen{nf.id, nf.text, nf.content})
		}
	}

	if len(toDelete) == 0 {
		s.emit(dupCtx, jobID, "info", "no duplicate or near-duplicate facts found")
		s.emit(dupCtx, jobID, "debug", fmt.Sprintf("duplicate check finished in %.2fs", time.Since(start).Seconds()))
		return
	}

	s.emit(dupCtx, jobID, "info", fmt.Sprintf("deleting %d duplicate facts...", len(toDelete)))

	if err := s.factRepository.BulkDelete(dupCtx, toDelete); err != nil {
		s.emit(dupCtx, jobID, "error", fmt.Sprintf("failed to delete duplicate facts: %v", err))
		return
	}

	s.emit(dupCtx, jobID, "info", fmt.Sprintf("deleted %d duplicate facts", len(toDelete)))
	s.emit(dupCtx, jobID, "debug", fmt.Sprintf("duplicate check finished in %.2fs", time.Since(start).Seconds()))
}

func (s *GCService) cleanupOldJobs(ctx context.Context, jobID string) {
	before := time.Now().Add(-s.logRetention)

	s.emit(ctx, jobID, "info", fmt.Sprintf("cleaning up gc jobs older than %s...", before.Format("2006-01-02")))

	oldJobs, err := s.gcJobRepository.FindOlderThan(ctx, before)
	if err != nil {
		s.emit(ctx, jobID, "error", fmt.Sprintf("failed to fetch old jobs: %v", err))
		return
	}

	if len(oldJobs) == 0 {
		s.emit(ctx, jobID, "info", "no old gc jobs to clean up")
		return
	}

	ids := make([]string, len(oldJobs))
	for i, j := range oldJobs {
		ids[i] = j.ID
	}

	if err := s.gcJobRepository.DeleteBatch(ctx, ids); err != nil {
		s.emit(ctx, jobID, "error", fmt.Sprintf("failed to delete old jobs: %v", err))
		return
	}

	s.emit(ctx, jobID, "info", fmt.Sprintf("deleted %d old gc jobs", len(ids)))
}
