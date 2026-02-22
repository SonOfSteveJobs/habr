package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func TestLogin_Success(t *testing.T) {
	user := testUser(t)

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	tokenRepo := &mockTokenRepo{
		saveFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return nil },
	}
	svc := newTestService(userRepo, tokenRepo)

	pair, err := svc.Login(context.Background(), "user@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}

	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}

	if !tokenRepo.saveCalled {
		t.Error("tokenRepo.Save was not called")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, model.ErrUserNotFound
		},
	}
	svc := newTestService(userRepo, &mockTokenRepo{})

	_, err := svc.Login(context.Background(), "noone@example.com", "password")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := testUser(t)

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	svc := newTestService(userRepo, &mockTokenRepo{})

	_, err := svc.Login(context.Background(), "user@example.com", "wrongpassword")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_TokenRepoError(t *testing.T) {
	user := testUser(t)
	redisErr := errors.New("redis connection refused")

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	tokenRepo := &mockTokenRepo{
		saveFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return redisErr },
	}
	svc := newTestService(userRepo, tokenRepo)

	_, err := svc.Login(context.Background(), "user@example.com", "correctpassword")
	if !errors.Is(err, redisErr) {
		t.Errorf("error = %v, want %v", err, redisErr)
	}
}
