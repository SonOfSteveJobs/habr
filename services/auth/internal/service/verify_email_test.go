package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func TestVerifyEmail_Success(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	userRepo := &mockUserRepo{
		confirmEmailFn: func(_ context.Context, id uuid.UUID) error {
			if id != userID {
				t.Errorf("unexpected userID: %v", id)
			}
			return nil
		},
	}

	verificationRepo := &mockVerificationRepo{
		validateFn: func(_ context.Context, code string, id uuid.UUID) error {
			if code != "123456" || id != userID {
				t.Error("unexpected args")
			}
			return nil
		},
	}

	svc := newTestServiceWithVerification(userRepo, &mockTokenRepo{}, verificationRepo)

	err := svc.VerifyEmail(context.Background(), userID, "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyEmail_InvalidCode(t *testing.T) {
	verificationRepo := &mockVerificationRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error {
			return model.ErrInvalidVerificationCode
		},
	}

	svc := newTestServiceWithVerification(&mockUserRepo{}, &mockTokenRepo{}, verificationRepo)

	err := svc.VerifyEmail(context.Background(), uuid.Must(uuid.NewV7()), "000000")
	if !errors.Is(err, model.ErrInvalidVerificationCode) {
		t.Errorf("error = %v, want ErrInvalidVerificationCode", err)
	}
}

func TestVerifyEmail_UserNotFound(t *testing.T) {
	verificationRepo := &mockVerificationRepo{
		validateFn: func(_ context.Context, _ string, _ uuid.UUID) error {
			return nil
		},
	}

	userRepo := &mockUserRepo{
		confirmEmailFn: func(_ context.Context, _ uuid.UUID) error {
			return model.ErrUserNotFound
		},
	}

	svc := newTestServiceWithVerification(userRepo, &mockTokenRepo{}, verificationRepo)

	err := svc.VerifyEmail(context.Background(), uuid.Must(uuid.NewV7()), "123456")
	if !errors.Is(err, model.ErrUserNotFound) {
		t.Errorf("error = %v, want ErrUserNotFound", err)
	}
}
