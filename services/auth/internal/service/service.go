package service

import (
	"context"
	"errors"
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
