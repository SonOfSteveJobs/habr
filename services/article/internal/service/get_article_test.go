package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func TestGetArticle_Success(t *testing.T) {
	articleID := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())
	expected := &model.Article{
		ID:        articleID,
		AuthorID:  authorID,
		Title:     "Test Title",
		Content:   "Test Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo := &mockArticleRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*model.Article, error) {
			if id != articleID {
				t.Errorf("id = %v, want %v", id, articleID)
			}
			return expected, nil
		},
	}
	svc := newTestService(repo)

	article, err := svc.GetArticle(context.Background(), articleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.ID != articleID {
		t.Errorf("article.ID = %v, want %v", article.ID, articleID)
	}

	if article.Title != "Test Title" {
		t.Errorf("article.Title = %q, want %q", article.Title, "Test Title")
	}

	if !repo.getByIDCalled {
		t.Error("repo.GetByID was not called")
	}
}

func TestGetArticle_NotFound(t *testing.T) {
	repo := &mockArticleRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Article, error) {
			return nil, model.ErrArticleNotFound
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetArticle(context.Background(), uuid.Must(uuid.NewV7()))
	if !errors.Is(err, model.ErrArticleNotFound) {
		t.Errorf("error = %v, want ErrArticleNotFound", err)
	}
}

func TestGetArticle_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockArticleRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Article, error) {
			return nil, repoErr
		},
	}
	svc := newTestService(repo)

	_, err := svc.GetArticle(context.Background(), uuid.Must(uuid.NewV7()))
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
