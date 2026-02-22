package service

import (
	"context"
	"errors"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func (s *Service) Login(ctx context.Context, email, password string) (*model.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.ErrInvalidCredentials
		}

		return nil, err
	}

	if err := user.VerifyPassword(password); err != nil {
		return nil, err
	}

	pair, err := model.NewTokenPair(user.ID, s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.Save(ctx, pair.RefreshToken, user.ID, s.refreshTTL); err != nil {
		return nil, err
	}

	return pair, nil
}
