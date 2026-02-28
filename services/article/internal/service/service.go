package service

import (
	"context"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
}

type TxManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	articleRepo ArticleRepository
	txManager   TxManager
}

func New(articleRepo ArticleRepository, txManager TxManager) *Service {
	return &Service{
		articleRepo: articleRepo,
		txManager:   txManager,
	}
}
