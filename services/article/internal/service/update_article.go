package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func (s *Service) UpdateArticle(ctx context.Context, id, authorID uuid.UUID, title, content *string) (*model.Article, error) {
	article := new(model.Article)
	if err := article.Update(id, authorID, title, content); err != nil {
		return nil, fmt.Errorf("update article model: %w", err)
	}

	if err := s.articleRepo.Update(ctx, article); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrArticleNotFound
		}
		return nil, fmt.Errorf("update article: %w", err)
	}

	if err := s.cacheRepo.Invalidate(ctx); err != nil {
		log := logger.Ctx(ctx)
		log.Warn().Err(err).Msg("cache invalidate failed")
	}

	return article, nil
}
