package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type mockArticleRepo struct {
	createFn     func(ctx context.Context, article *model.Article) error
	createCalled bool

	listFn     func(ctx context.Context, cursor string, limit int) (*model.ArticlePage, error)
	listCalled bool

	getByIDFn     func(ctx context.Context, id uuid.UUID) (*model.Article, error)
	getByIDCalled bool

	updateFn     func(ctx context.Context, article *model.Article) error
	updateCalled bool

	deleteFn     func(ctx context.Context, id, authorID uuid.UUID) error
	deleteCalled bool
}

func (m *mockArticleRepo) Create(ctx context.Context, article *model.Article) error {
	m.createCalled = true
	return m.createFn(ctx, article)
}

func (m *mockArticleRepo) List(ctx context.Context, cursor string, limit int) (*model.ArticlePage, error) {
	m.listCalled = true
	return m.listFn(ctx, cursor, limit)
}

func (m *mockArticleRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	m.getByIDCalled = true
	return m.getByIDFn(ctx, id)
}

func (m *mockArticleRepo) Update(ctx context.Context, article *model.Article) error {
	m.updateCalled = true
	return m.updateFn(ctx, article)
}

func (m *mockArticleRepo) Delete(ctx context.Context, id, authorID uuid.UUID) error {
	m.deleteCalled = true
	return m.deleteFn(ctx, id, authorID)
}

type mockCacheRepo struct {
	getFn        func(ctx context.Context) (*model.ArticlePage, error)
	setFn        func(ctx context.Context, page *model.ArticlePage) error
	invalidateFn func(ctx context.Context) error
}

func (m *mockCacheRepo) Get(ctx context.Context) (*model.ArticlePage, error) {
	return m.getFn(ctx)
}

func (m *mockCacheRepo) Set(ctx context.Context, page *model.ArticlePage) error {
	return m.setFn(ctx, page)
}

func (m *mockCacheRepo) Invalidate(ctx context.Context) error {
	return m.invalidateFn(ctx)
}

type mockTxManager struct{}

func (m *mockTxManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func defaultCacheRepo() *mockCacheRepo {
	return &mockCacheRepo{
		getFn:        func(_ context.Context) (*model.ArticlePage, error) { return nil, nil },
		setFn:        func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error { return nil },
	}
}

func newTestService(repo *mockArticleRepo) *Service {
	return New(repo, defaultCacheRepo(), &mockTxManager{})
}

func newTestServiceWithCache(repo *mockArticleRepo, cache *mockCacheRepo) *Service {
	return New(repo, cache, &mockTxManager{})
}
