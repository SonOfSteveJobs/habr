package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestLogout_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tokenRepo := &mockTokenRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	err := svc.Logout(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !tokenRepo.deleteCalled {
		t.Error("tokenRepo.Delete was not called")
	}
}

func TestLogout_DeleteError(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	redisErr := errors.New("redis error")

	tokenRepo := &mockTokenRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error { return redisErr },
	}
	svc := newTestService(&mockUserRepo{}, tokenRepo)

	err := svc.Logout(context.Background(), userID)
	if !errors.Is(err, redisErr) {
		t.Errorf("error = %v, want %v", err, redisErr)
	}
}
