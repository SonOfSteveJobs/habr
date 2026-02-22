package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func (s *Service) RefreshToken(ctx context.Context, userID uuid.UUID, oldRefreshToken string) (*model.TokenPair, error) {
	if err := s.tokenRepo.Validate(ctx, oldRefreshToken, userID); err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Delete(ctx, userID); err != nil {
		return nil, err
	}

	pair, err := model.NewTokenPair(userID, s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Save(ctx, pair.RefreshToken, userID, s.refreshTTL); err != nil {
		return nil, err
	}

	return pair, nil
}
