package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
	List(ctx context.Context, cursor string, limit int) (*model.ArticlePage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error)
	Update(ctx context.Context, article *model.Article) error
	Delete(ctx context.Context, id, authorId uuid.UUID) error
}

type CacheRepository interface {
	Get(ctx context.Context) (*model.ArticlePage, error)
	Set(ctx context.Context, page *model.ArticlePage) error
	Invalidate(ctx context.Context) error
}

type TxManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	articleRepo ArticleRepository
	cacheRepo   CacheRepository
	txManager   TxManager
}

func New(articleRepo ArticleRepository, cacheRepo CacheRepository, txManager TxManager) *Service {
	return &Service{
		articleRepo: articleRepo,
		cacheRepo:   cacheRepo,
		txManager:   txManager,
	}
}
