package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func strPtr(s string) *string { return &s }

func TestUpdateArticle_Success(t *testing.T) {
	articleID := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())

	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, a *model.Article) error {
			if a.ID != articleID {
				t.Errorf("article.ID = %v, want %v", a.ID, articleID)
			}
			if a.AuthorID != authorID {
				t.Errorf("article.AuthorID = %v, want %v", a.AuthorID, authorID)
			}
			if a.Title != "Updated Title" {
				t.Errorf("article.Title = %q, want %q", a.Title, "Updated Title")
			}
			if a.Content != "Updated Content" {
				t.Errorf("article.Content = %q, want %q", a.Content, "Updated Content")
			}
			return nil
		},
	}
	svc := newTestService(repo)

	article, err := svc.UpdateArticle(context.Background(), articleID, authorID, strPtr("Updated Title"), strPtr("Updated Content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.ID != articleID {
		t.Errorf("article.ID = %v, want %v", article.ID, articleID)
	}

	if article.Title != "Updated Title" {
		t.Errorf("article.Title = %q, want %q", article.Title, "Updated Title")
	}

	if !repo.updateCalled {
		t.Error("repo.Update was not called")
	}
}

func TestUpdateArticle_TitleOnly(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, a *model.Article) error {
			if a.Title != "New Title" {
				t.Errorf("article.Title = %q, want %q", a.Title, "New Title")
			}
			if a.Content != "" {
				t.Errorf("article.Content = %q, want empty (nil content)", a.Content)
			}
			return nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr("New Title"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.updateCalled {
		t.Error("repo.Update was not called")
	}
}

func TestUpdateArticle_ContentOnly(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, a *model.Article) error {
			if a.Title != "" {
				t.Errorf("article.Title = %q, want empty (nil title)", a.Title)
			}
			if a.Content != "New Content" {
				t.Errorf("article.Content = %q, want %q", a.Content, "New Content")
			}
			return nil
		},
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil, strPtr("New Content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.updateCalled {
		t.Error("repo.Update was not called")
	}
}

func TestUpdateArticle_EmptyTitle(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr(""), strPtr("content"))
	if !errors.Is(err, model.ErrInvalidTitle) {
		t.Errorf("error = %v, want ErrInvalidTitle", err)
	}

	if repo.updateCalled {
		t.Error("repo.Update was called, want skipped on invalid title")
	}
}

func TestUpdateArticle_TitleTooLong(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	longTitle := strings.Repeat("a", 256)
	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr(longTitle), nil)
	if !errors.Is(err, model.ErrInvalidTitle) {
		t.Errorf("error = %v, want ErrInvalidTitle", err)
	}

	if repo.updateCalled {
		t.Error("repo.Update was called, want skipped on long title")
	}
}

func TestUpdateArticle_EmptyContent(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil, strPtr(""))
	if !errors.Is(err, model.ErrInvalidContent) {
		t.Errorf("error = %v, want ErrInvalidContent", err)
	}

	if repo.updateCalled {
		t.Error("repo.Update was called, want skipped on empty content")
	}
}

func TestUpdateArticle_ContentTooLong(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	longContent := strings.Repeat("a", 50001)
	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), nil, strPtr(longContent))
	if !errors.Is(err, model.ErrInvalidContent) {
		t.Errorf("error = %v, want ErrInvalidContent", err)
	}

	if repo.updateCalled {
		t.Error("repo.Update was called, want skipped on long content")
	}
}

func TestUpdateArticle_NotFound(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error {
			return sql.ErrNoRows
		},
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr("title"), nil)
	if !errors.Is(err, model.ErrArticleNotFound) {
		t.Errorf("error = %v, want ErrArticleNotFound", err)
	}
}

func TestUpdateArticle_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return repoErr },
	}
	svc := newTestService(repo)

	_, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr("title"), nil)
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}

func TestUpdateArticle_CacheInvalidateErrorNonFatal(t *testing.T) {
	repo := &mockArticleRepo{
		updateFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	cache := &mockCacheRepo{
		getFn: func(_ context.Context) (*model.ArticlePage, error) { return nil, nil },
		setFn: func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error {
			return errors.New("redis connection refused")
		},
	}
	svc := newTestServiceWithCache(repo, cache)

	article, err := svc.UpdateArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), strPtr("title"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article == nil {
		t.Error("article = nil, want non-nil")
	}
}
