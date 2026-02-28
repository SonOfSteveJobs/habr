package service

import (
	"context"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type mockArticleRepo struct {
	createFn     func(ctx context.Context, article *model.Article) error
	createCalled bool
}

func (m *mockArticleRepo) Create(ctx context.Context, article *model.Article) error {
	m.createCalled = true
	return m.createFn(ctx, article)
}

type mockTxManager struct{}

func (m *mockTxManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newTestService(repo *mockArticleRepo) *Service {
	return New(repo, &mockTxManager{})
}
