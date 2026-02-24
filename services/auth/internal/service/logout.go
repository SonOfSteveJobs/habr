package service

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.tokenRepo.Delete(ctx, userID)
}
