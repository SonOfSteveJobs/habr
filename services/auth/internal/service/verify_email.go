package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) VerifyEmail(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.verificationRepo.Validate(ctx, code, userID); err != nil {
		return fmt.Errorf("validate verification code: %w", err)
	}

	if err := s.userRepo.ConfirmEmail(ctx, userID); err != nil {
		return fmt.Errorf("confirm email: %w", err)
	}

	// ну не удалили и ладно, по ttl удалится
	_ = s.verificationRepo.Delete(ctx, userID)

	return nil
}
