package service

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/repository"
	"KnowLedger/pkg/dto"
	"KnowLedger/pkg/utils"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type FunFactService struct {
	db             *gorm.DB
	factRepository *repository.FactRepository
	tagRepository  *repository.TagRepository
	mediaService   *MediaService
}

func NewFactService(db *gorm.DB, factRepo *repository.FactRepository, tagRepo *repository.TagRepository, service *MediaService) *FunFactService {
	return &FunFactService{
		db:             db,
		factRepository: factRepo,
		tagRepository:  tagRepo,
		mediaService:   service,
	}
}

func (s *FunFactService) DeleteFact(ctx context.Context, id string) error {
	if err := s.factRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete fact: %w", err)
	}
	return nil
}

func (s *FunFactService) CreateFact(ctx context.Context, fact *dto.PostCreateFunFactRequest) error {
	tagNames := utils.FormatTagsStrToSlice(fact.Tags)
	content := fact.Content
	status := fact.Status
	sourceURL := fact.SourceURL
	mediaKey := fact.MediaKey

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		factRepo := s.factRepository.WithTx(tx)
		tagRepo := s.tagRepository.WithTx(tx)

		var media *model.MediaItem
		if mediaKey != "" {
			m, err := s.mediaService.GetMediaDetails(ctx, mediaKey)
			if err != nil {
				return fmt.Errorf("failed to create fun fact, invalid media key: %w", err)
			}
			media = m
		}

		existingTags, err := tagRepo.FindTagsByNames(ctx, tagNames)
		if err != nil {
			return fmt.Errorf("failed to fetch existing tags: %w", err)
		}

		existingTagsMap := make(map[string]model.Tag)
		for _, t := range existingTags {
			existingTagsMap[t.Name] = *t
		}

		var finalTags []model.Tag
		var newTagNames []string
		var newTags []model.Tag

		for _, name := range tagNames {
			if tag, found := existingTagsMap[name]; found {
				finalTags = append(finalTags, tag)
			} else {
				newTags = append(newTags, model.Tag{Name: name})
				newTagNames = append(newTagNames, name)
			}
		}

		if len(newTags) > 0 {
			if err := tagRepo.CreateBulkIgnoreConflict(ctx, &newTags); err != nil {
				return fmt.Errorf("failed to create new tags: %w", err)
			}

			insertedTags, err := tagRepo.FindTagsByNames(ctx, newTagNames)
			if err != nil {
				return fmt.Errorf("failed to re-fetch tags after upsert: %w", err)
			}
			for _, t := range insertedTags {
				finalTags = append(finalTags, *t)
			}
		}

		newFact := model.Fact{
			Content:   content,
			Status:    status,
			SourceURL: sourceURL,
			Tags:      finalTags,
			Media:     media,
		}

		return factRepo.Create(ctx, &newFact)
	})

	if err != nil {
		return fmt.Errorf("failed to create fact: %w", err)
	}
	return nil
}

const tagBatchSize = 200
const factBatchSize = 150

// CreateFactBulk adds fact data in bulk with incremental batch queries.
func (s *FunFactService) CreateFactBulk(ctx context.Context, facts []*dto.PostCreateFunFactRequest) (failed int, err error) {
	if len(facts) == 0 {
		return 0, nil
	}

	tagNameSet := make(map[string]struct{})
	for _, f := range facts {
		for _, name := range utils.FormatTagsStrToSlice(f.Tags) {
			tagNameSet[name] = struct{}{}
		}
	}
	allTagNames := make([]string, 0, len(tagNameSet))
	for name := range tagNameSet {
		allTagNames = append(allTagNames, name)
	}

	var failCount int
	var batchErrs error

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		factRepo := s.factRepository.WithTx(tx)
		tagRepo := s.tagRepository.WithTx(tx)

		existingTags, err := tagRepo.FindTagsByNames(ctx, allTagNames)
		if err != nil {
			return fmt.Errorf("failed to fetch existing tags: %w", err)
		}

		tagMap := make(map[string]model.Tag, len(existingTags))
		for _, t := range existingTags {
			tagMap[t.Name] = *t
		}

		var newTags []model.Tag
		for _, name := range allTagNames {
			if _, exists := tagMap[name]; !exists {
				newTags = append(newTags, model.Tag{Name: name})
			}
		}

		if len(newTags) > 0 {
			if err := repository.InsertInBatches(ctx, newTags, tagBatchSize, func(batch []model.Tag) error {
				return tagRepo.CreateBulkIgnoreConflict(ctx, &batch)
			}); err != nil {
				return fmt.Errorf("failed to upsert new tags: %w", err)
			}

			newTagNames := make([]string, len(newTags))
			for i, t := range newTags {
				newTagNames[i] = t.Name
			}
			insertedTags, err := tagRepo.FindTagsByNames(ctx, newTagNames)
			if err != nil {
				return fmt.Errorf("failed to re-fetch new tags: %w", err)
			}
			for _, t := range insertedTags {
				tagMap[t.Name] = *t
			}
		}

		var newFacts []model.Fact
		for _, f := range facts {
			var media *model.MediaItem
			if f.MediaKey != "" {
				m, err := s.mediaService.GetMediaDetails(ctx, f.MediaKey)
				if err != nil {
					failCount++
					batchErrs = errors.Join(batchErrs, fmt.Errorf(
						"skipped fact (content: %.30q): invalid media key: %w",
						f.Content, err,
					))
					continue
				}
				media = m
			}

			tagNames := utils.FormatTagsStrToSlice(f.Tags)
			factTags := make([]model.Tag, 0, len(tagNames))
			for _, name := range tagNames {
				if tag, ok := tagMap[name]; ok {
					factTags = append(factTags, tag)
				}
			}

			newFacts = append(newFacts, model.Fact{
				Content:   f.Content,
				Status:    model.FactStatusDraft,
				SourceURL: f.SourceURL,
				Tags:      factTags,
				Media:     media,
			})
		}

		batchStart := 0
		if err := repository.InsertInBatches(ctx, newFacts, factBatchSize, func(batch []model.Fact) error {
			if err := factRepo.CreateBulk(ctx, batch); err != nil {
				failCount += len(batch)
				batchErrs = errors.Join(batchErrs, fmt.Errorf(
					"batch [%d..%d] failed: %w",
					batchStart, batchStart+len(batch)-1, err,
				))
			}
			batchStart += len(batch)
			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		// if the process reaches this part, it means there is a very fatal error.
		return len(facts), fmt.Errorf("transaction failed: %w", txErr)
	}

	return failCount, batchErrs
}

func (s *FunFactService) GetFunFactStats(ctx context.Context) (*model.FunFactStats, error) {
	g, ctx := errgroup.WithContext(ctx)

	var facts int64
	var tags int64

	g.Go(func() error {
		factsCount, err := s.factRepository.Count(ctx)
		if err != nil {
			return err
		}
		facts = factsCount
		return nil
	})
	g.Go(func() error {
		tagsCount, err := s.tagRepository.Count(ctx)
		if err != nil {
			return err
		}
		tags = tagsCount
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to get fun fact statistics: %w", err)
	}

	return &model.FunFactStats{
		FunFacts: facts,
		Tags:     tags,
	}, nil
}

func (s *FunFactService) GetOneRandomFunFact(ctx context.Context) (*model.Fact, error) {
	fact, err := s.factRepository.GetRandomOne(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get one random fun fact: %w", err)
	}
	return fact, nil
}

func (s *FunFactService) GetOneFunFact(ctx context.Context, id string) (*model.Fact, error) {
	fact, err := s.factRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get find fun fact: %w", err)
	}
	return fact, nil
}

func (s *FunFactService) GetFacts(ctx context.Context, params *dto.ListFactsParams) (*model.Paginated[*model.Fact], error) {
	facts, err := s.factRepository.GetFacts(ctx, model.ListFactsParams{
		Page:    params.Page,
		Limit:   params.Limit,
		Search:  params.Search,
		Status:  params.Status,
		SortBy:  params.SortBy,
		SortDir: params.SortDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get facts: %w", err)
	}
	return facts, nil
}

func (s *FunFactService) UpdateFunFact(ctx context.Context, id string, request *dto.PostEditFunFactRequest) (*model.Fact, error) {
	updatedFact := &model.Fact{
		Content:   request.Content,
		Status:    request.Status,
		SourceURL: request.SourceURL,
	}
	tags := utils.FormatTagsStrToSlice(request.Tags)

	if request.MediaKey != "" {
		media, err := s.mediaService.GetMediaDetails(ctx, request.MediaKey)
		if err != nil {
			return nil, fmt.Errorf("failed to update fun fact, invalid media data: %w", err)
		}
		updatedFact.Media = media
	}

	fact, err := s.factRepository.Update(ctx, id, updatedFact, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to update fun fact: %w", err)
	}

	return fact, nil
}

func (s *FunFactService) GetTags(ctx context.Context, params *dto.ListTagsParams) (*model.Paginated[*model.Tag], error) {
	tags, err := s.tagRepository.GetTags(ctx, model.ListTagsParams{
		Page:  params.Page,
		Limit: params.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	return tags, nil
}

func (s *FunFactService) DeleteTag(ctx context.Context, id string) error {
	if err := s.tagRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	return nil
}

func (s *FunFactService) GetTagSuggestions(ctx context.Context, q string) []string {
	limit := 10
	tags, err := s.tagRepository.SearchTagByName(ctx, q, limit)
	if err != nil {
		zap.L().Error("failed to get tag suggestions", zap.String("q", q), zap.Error(err))
		return []string{}
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}

	return names
}
