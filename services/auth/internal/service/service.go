package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

type Service struct {
	userRepo UserRepository
}

func New(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

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
