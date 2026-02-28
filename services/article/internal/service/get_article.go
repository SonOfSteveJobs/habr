package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func (s *Service) GetArticle(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get article: %w", err)
	}

	return article, nil
}
