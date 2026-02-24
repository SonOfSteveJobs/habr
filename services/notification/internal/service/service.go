package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/services/notification/internal/model"
)

type EventRepository interface {
	MarkProcessed(ctx context.Context, eventID uuid.UUID) (bool, error)
}

type TxManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type EmailSender interface {
	Send(ctx context.Context, event model.UserRegisteredEvent) error
}

type Service struct {
	eventRepo   EventRepository
	txManager   TxManager
	emailSender EmailSender
	eventTTL    time.Duration
}

func New(
	eventRepo EventRepository,
	txManager TxManager,
	emailSender EmailSender,
	eventTTL time.Duration,
) *Service {
	return &Service{
		eventRepo:   eventRepo,
		txManager:   txManager,
		emailSender: emailSender,
		eventTTL:    eventTTL,
	}
}
