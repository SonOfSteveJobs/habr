package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func TestCreateArticle_Success(t *testing.T) {
	repo := &mockArticleRepo{
		createFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	authorID := uuid.Must(uuid.NewV7())
	article, err := svc.CreateArticle(context.Background(), authorID, "Test Title", "Test Content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.ID == uuid.Nil {
		t.Error("article.ID = uuid.Nil, want valid UUID")
	}

	if article.AuthorID != authorID {
		t.Errorf("article.AuthorID = %v, want %v", article.AuthorID, authorID)
	}

	if article.Title != "Test Title" {
		t.Errorf("article.Title = %q, want %q", article.Title, "Test Title")
	}

	if article.Content != "Test Content" {
		t.Errorf("article.Content = %q, want %q", article.Content, "Test Content")
	}

	if !repo.createCalled {
		t.Error("repo.Create was not called")
	}
}

func TestCreateArticle_EmptyTitle(t *testing.T) {
	repo := &mockArticleRepo{
		createFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	_, err := svc.CreateArticle(context.Background(), uuid.Must(uuid.NewV7()), "", "content")
	if !errors.Is(err, model.ErrInvalidTitle) {
		t.Errorf("error = %v, want ErrInvalidTitle", err)
	}

	if repo.createCalled {
		t.Error("repo.Create was called, want skipped on invalid title")
	}
}

func TestCreateArticle_TitleTooLong(t *testing.T) {
	repo := &mockArticleRepo{
		createFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	longTitle := strings.Repeat("a", 256)
	_, err := svc.CreateArticle(context.Background(), uuid.Must(uuid.NewV7()), longTitle, "content")
	if !errors.Is(err, model.ErrInvalidTitle) {
		t.Errorf("error = %v, want ErrInvalidTitle", err)
	}

	if repo.createCalled {
		t.Error("repo.Create was called, want skipped on long title")
	}
}

func TestCreateArticle_EmptyContent(t *testing.T) {
	repo := &mockArticleRepo{
		createFn: func(_ context.Context, _ *model.Article) error { return nil },
	}
	svc := newTestService(repo)

	_, err := svc.CreateArticle(context.Background(), uuid.Must(uuid.NewV7()), "title", "")
	if !errors.Is(err, model.ErrInvalidContent) {
		t.Errorf("error = %v, want ErrInvalidContent", err)
	}

	if repo.createCalled {
		t.Error("repo.Create was called, want skipped on empty content")
	}
}

func TestCreateArticle_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockArticleRepo{
		createFn: func(_ context.Context, _ *model.Article) error { return repoErr },
	}
	svc := newTestService(repo)

	_, err := svc.CreateArticle(context.Background(), uuid.Must(uuid.NewV7()), "title", "content")
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
