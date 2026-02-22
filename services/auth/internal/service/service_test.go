package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type mockUserRepo struct {
	createFn func(ctx context.Context, user *model.User) error
	called   bool
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	m.called = true
	return m.createFn(ctx, user)
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error {
			return nil
		},
	}
	svc := New(repo)

	id, err := svc.Register(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id == uuid.Nil {
		t.Error("returned uuid.Nil, want valid UUID")
	}

	if !repo.called {
		t.Error("repo.Create was not called")
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error {
			return model.ErrEmailAlreadyExists
		},
	}
	svc := New(repo)

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, model.ErrEmailAlreadyExists) {
		t.Errorf("error = %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegister_InvalidEmail_RepoNotCalled(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error {
			return nil
		},
	}
	svc := New(repo)

	_, err := svc.Register(context.Background(), "invalid", "password")
	if !errors.Is(err, model.ErrInvalidEmail) {
		t.Errorf("error = %v, want ErrInvalidEmail", err)
	}

	if repo.called {
		t.Error("repo.Create was called, want skipped on invalid email")
	}
}

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("connection refused")
	repo := &mockUserRepo{
		createFn: func(_ context.Context, _ *model.User) error {
			return repoErr
		},
	}
	svc := New(repo)

	_, err := svc.Register(context.Background(), "user@example.com", "password")
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}
}
