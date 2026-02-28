package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func (s *Service) CreateArticle(ctx context.Context, authorID uuid.UUID, title, content string) (*model.Article, error) {
	article, err := model.NewArticle(authorID, title, content)
	if err != nil {
		return nil, fmt.Errorf("create article model: %w", err)
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("save article: %w", err)
	}

	if err := s.cacheRepo.Invalidate(ctx); err != nil {
		log := logger.Ctx(ctx)
		log.Warn().Err(err).Msg("cache invalidate failed")
	}

	return article, nil
}
