package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return nil },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	id, err := svc.Register(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id == uuid.Nil {
		t.Error("returned uuid.Nil, want valid UUID")
	}

	if !repo.createCalled {
		t.Error("repo.Create was not called")
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return model.ErrEmailAlreadyExists },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, model.ErrEmailAlreadyExists) {
		t.Errorf("error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegister_InvalidEmail_RepoNotCalled(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return nil },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "invalid", "password")
	if !errors.Is(err, model.ErrInvalidEmail) {
		t.Errorf("error = %v, want ErrInvalidEmail", err)
	}

	if repo.createCalled {
		t.Error("repo.Create was called, want skipped on invalid email")
	}
}

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error { return repoErr },
	}
	svc := newTestService(repo, &mockTokenRepo{})

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
