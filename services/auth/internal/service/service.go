package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type TokenRepository interface {
	Save(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error
	Validate(ctx context.Context, refreshToken string, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	userRepo   UserRepository
	tokenRepo  TokenRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}
