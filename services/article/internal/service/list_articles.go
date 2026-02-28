package service

import (
	"context"
	"fmt"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

const defaultLimit = 20

func (s *Service) ListArticles(ctx context.Context, cursor string, limit int32) (*model.ArticlePage, error) {
	l := int(limit)
	if l <= 0 {
		l = defaultLimit
	}

	isFirstPage := cursor == ""

	if isFirstPage {
		page, err := s.cacheRepo.Get(ctx)
		if err != nil {
			log := logger.Ctx(ctx)
			log.Warn().Err(err).Msg("cache get failed")
		}
		if page != nil {
			return page, nil
		}
	}

	page, err := s.articleRepo.List(ctx, cursor, l)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}

	if isFirstPage {
		if err := s.cacheRepo.Set(ctx, page); err != nil {
			log := logger.Ctx(ctx)
			log.Warn().Err(err).Msg("cache set failed")
		}
	}

	return page, nil
}
