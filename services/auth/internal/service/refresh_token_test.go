package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func TestRefreshToken_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tokenRepo := &mockTokenRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error { return nil },
		deleteFn:   func(_ context.Context, _ uuid.UUID) error { return nil },
		saveFn:     func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return nil },
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	pair, err := svc.RefreshToken(context.Background(), userID, "old-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}

	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}

	if !tokenRepo.deleteCalled {
		t.Error("tokenRepo.Delete was not called")
	}

	if !tokenRepo.saveCalled {
		t.Error("tokenRepo.Save was not called")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tokenRepo := &mockTokenRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error {
			return model.ErrInvalidRefreshToken
		},
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	_, err := svc.RefreshToken(context.Background(), userID, "bad-token")
	if !errors.Is(err, model.ErrInvalidRefreshToken) {
		t.Errorf("error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefreshToken_DeleteError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	redisErr := errors.New("redis error")

	tokenRepo := &mockTokenRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error { return nil },
		deleteFn:   func(_ context.Context, _ uuid.UUID) error { return redisErr },
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	_, err := svc.RefreshToken(context.Background(), userID, "old-token")
	if !errors.Is(err, redisErr) {
		t.Errorf("error = %v, want %v", err, redisErr)
	}
}

func TestRefreshToken_SaveError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	redisErr := errors.New("redis error")

	tokenRepo := &mockTokenRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error { return nil },
		deleteFn:   func(_ context.Context, _ uuid.UUID) error { return nil },
		saveFn:     func(_ context.Context, _ string, _ uuid.UUID, _ time.Duration) error { return redisErr },
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	_, err := svc.RefreshToken(context.Background(), userID, "old-token")
	if !errors.Is(err, redisErr) {
		t.Errorf("error = %v, want %v", err, redisErr)
	}
}
