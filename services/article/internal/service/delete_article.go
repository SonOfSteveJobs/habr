package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

func (s *Service) DeleteArticle(ctx context.Context, id, authorID uuid.UUID) error {
	if err := s.articleRepo.Delete(ctx, id, authorID); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}

	if err := s.cacheRepo.Invalidate(ctx); err != nil {
		log := logger.Ctx(ctx)
		log.Warn().Err(err).Msg("cache invalidate failed")
	}

	return nil
}
