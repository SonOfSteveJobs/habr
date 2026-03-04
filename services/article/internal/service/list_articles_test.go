package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func testArticlePage() *model.ArticlePage {
	return &model.ArticlePage{
		Articles: []*model.Article{
			{
				ID:        uuid.Must(uuid.NewV7()),
				AuthorID:  uuid.Must(uuid.NewV7()),
				Title:     "Test Article",
				Content:   "Test Content",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		NextCursor: "next",
	}
}

func TestListArticles_CacheHit(t *testing.T) {
	cachedPage := testArticlePage()

	repo := &mockArticleRepo{
		listFn: func(_ context.Context, _ string, _ int) (*model.ArticlePage, error) {
			t.Error("repo.List was called, want cache hit")
			return nil, nil
		},
	}
	cache := &mockCacheRepo{
		getFn:        func(_ context.Context) (*model.ArticlePage, error) { return cachedPage, nil },
		setFn:        func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error { return nil },
	}
	svc := newTestServiceWithCache(repo, cache)

	page, err := svc.ListArticles(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(page.Articles) != 1 {
		t.Errorf("articles count = %d, want 1", len(page.Articles))
	}

	if page.Articles[0].Title != "Test Article" {
		t.Errorf("title = %q, want %q", page.Articles[0].Title, "Test Article")
	}
}

func TestListArticles_CacheMiss(t *testing.T) {
	dbPage := testArticlePage()
	var setCalled bool

	repo := &mockArticleRepo{
		listFn: func(_ context.Context, _ string, _ int) (*model.ArticlePage, error) {
			return dbPage, nil
		},
	}
	cache := &mockCacheRepo{
		getFn: func(_ context.Context) (*model.ArticlePage, error) { return nil, nil },
		setFn: func(_ context.Context, _ *model.ArticlePage) error {
			setCalled = true
			return nil
		},
		invalidateFn: func(_ context.Context) error { return nil },
	}
	svc := newTestServiceWithCache(repo, cache)

	page, err := svc.ListArticles(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.listCalled {
		t.Error("repo.List was not called")
	}

	if !setCalled {
		t.Error("cache.Set was not called on first page")
	}

	if len(page.Articles) != 1 {
		t.Errorf("articles count = %d, want 1", len(page.Articles))
	}
}

func TestListArticles_CacheErrorFallbackToDB(t *testing.T) {
	dbPage := testArticlePage()

	repo := &mockArticleRepo{
		listFn: func(_ context.Context, _ string, _ int) (*model.ArticlePage, error) {
			return dbPage, nil
		},
	}
	cache := &mockCacheRepo{
		getFn: func(_ context.Context) (*model.ArticlePage, error) {
			return nil, errors.New("redis connection refused")
		},
		setFn:        func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error { return nil },
	}
	svc := newTestServiceWithCache(repo, cache)

	page, err := svc.ListArticles(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.listCalled {
		t.Error("repo.List was not called after cache error")
	}

	if len(page.Articles) != 1 {
		t.Errorf("articles count = %d, want 1", len(page.Articles))
	}
}

func TestListArticles_WithCursorSkipsCache(t *testing.T) {
	dbPage := testArticlePage()
	dbPage.NextCursor = ""
	var getCalled bool

	repo := &mockArticleRepo{
		listFn: func(_ context.Context, cursor string, _ int) (*model.ArticlePage, error) {
			if cursor != "some-cursor" {
				t.Errorf("cursor = %q, want %q", cursor, "some-cursor")
			}
			return dbPage, nil
		},
	}
	cache := &mockCacheRepo{
		getFn: func(_ context.Context) (*model.ArticlePage, error) {
			getCalled = true
			return nil, nil
		},
		setFn:        func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error { return nil },
	}
	svc := newTestServiceWithCache(repo, cache)

	_, err := svc.ListArticles(context.Background(), "some-cursor", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if getCalled {
		t.Error("cache.Get was called with cursor, want skipped")
	}
}

func TestListArticles_DefaultLimit(t *testing.T) {
	repo := &mockArticleRepo{
		listFn: func(_ context.Context, _ string, limit int) (*model.ArticlePage, error) {
			if limit != defaultLimit {
				t.Errorf("limit = %d, want %d", limit, defaultLimit)
			}
			return &model.ArticlePage{}, nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.ListArticles(context.Background(), "cursor", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListArticles_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockArticleRepo{
		listFn: func(_ context.Context, _ string, _ int) (*model.ArticlePage, error) {
			return nil, repoErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.ListArticles(context.Background(), "cursor", 20)
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
