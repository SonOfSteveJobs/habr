package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

func (s *Service) Register(ctx context.Context, email, password string) (uuid.UUID, error) {
	user, err := model.NewUser(email, password)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}
