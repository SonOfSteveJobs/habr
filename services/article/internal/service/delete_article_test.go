package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

func TestDeleteArticle_Success(t *testing.T) {
	articleID := uuid.Must(uuid.NewV7())
	authorID := uuid.Must(uuid.NewV7())

	repo := &mockArticleRepo{
		deleteFn: func(_ context.Context, id, authID uuid.UUID) error {
			if id != articleID {
				t.Errorf("id = %v, want %v", id, articleID)
			}
			if authID != authorID {
				t.Errorf("authorID = %v, want %v", authID, authorID)
			}
			return nil
		},
	}
	svc := newTestService(repo)

	err := svc.DeleteArticle(context.Background(), articleID, authorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.deleteCalled {
		t.Error("repo.Delete was not called")
	}
}

func TestDeleteArticle_NotFound(t *testing.T) {
	repo := &mockArticleRepo{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error {
			return model.ErrArticleNotFound
		},
	}
	svc := newTestService(repo)

	err := svc.DeleteArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))
	if !errors.Is(err, model.ErrArticleNotFound) {
		t.Errorf("error = %v, want ErrArticleNotFound", err)
	}
}

func TestDeleteArticle_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockArticleRepo{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return repoErr },
	}
	svc := newTestService(repo)

	err := svc.DeleteArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}

func TestDeleteArticle_CacheInvalidateErrorNonFatal(t *testing.T) {
	repo := &mockArticleRepo{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}
	cache := &mockCacheRepo{
		getFn: func(_ context.Context) (*model.ArticlePage, error) { return nil, nil },
		setFn: func(_ context.Context, _ *model.ArticlePage) error { return nil },
		invalidateFn: func(_ context.Context) error {
			return errors.New("redis connection refused")
		},
	}
	svc := newTestServiceWithCache(repo, cache)

	err := svc.DeleteArticle(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
