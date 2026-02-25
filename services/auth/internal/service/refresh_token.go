package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func (s *Service) RefreshToken(ctx context.Context, userID uuid.UUID, oldRefreshToken string) (*model.TokenPair, error) {
	if err := s.tokenRepo.Validate(ctx, oldRefreshToken, userID); err != nil {
		return nil, fmt.Errorf("token validate error: %w", err)
	}

	if err := s.tokenRepo.Delete(ctx, userID); err != nil {
		return nil, fmt.Errorf("token delete error: %w", err)
	}

	pair, err := model.NewTokenPair(userID, s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create token pair: %w", err)
	}

	if err := s.tokenRepo.Save(ctx, pair.RefreshToken, userID, s.refreshTTL); err != nil {
		return nil, fmt.Errorf("token save error: %w", err)
	}

	return pair, nil
}
