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
	ConfirmEmail(ctx context.Context, userID uuid.UUID) error
}

type TokenRepository interface {
	Save(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error
	Validate(ctx context.Context, refreshToken string, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type VerificationCodeRepository interface {
	Save(ctx context.Context, code string, userID uuid.UUID, ttl time.Duration) error
	Validate(ctx context.Context, code string, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type OutboxRepository interface {
	Insert(ctx context.Context, event model.OutboxEvent) error
}

type TxManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	userRepo         UserRepository
	tokenRepo        TokenRepository
	verificationRepo VerificationCodeRepository
	outboxRepo       OutboxRepository
	txManager        TxManager
	jwtSecret        string
	kafkaTopic       string
	accessTTL        time.Duration
	refreshTTL       time.Duration
	verificationTTL  time.Duration
}

func New(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	verificationRepo VerificationCodeRepository,
	outboxRepo OutboxRepository,
	txManager TxManager,
	jwtSecret string,
	kafkaTopic string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	verificationTTL time.Duration,
) *Service {
	return &Service{
		userRepo:         userRepo,
		tokenRepo:        tokenRepo,
		verificationRepo: verificationRepo,
		outboxRepo:       outboxRepo,
		txManager:        txManager,
		jwtSecret:        jwtSecret,
		kafkaTopic:       kafkaTopic,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
		verificationTTL:  verificationTTL,
	}
}
