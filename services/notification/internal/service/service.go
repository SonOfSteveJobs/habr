package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventRepository interface {
	MarkProcessed(ctx context.Context, eventID uuid.UUID) (bool, error)
	DeleteOld(ctx context.Context) error
}

type TxManager interface {
	Wrap(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	eventRepo EventRepository
	txManager TxManager
	eventTTL  time.Duration
}

func New(eventRepo EventRepository, txManager TxManager, eventTTL time.Duration) *Service {
	return &Service{
		eventRepo: eventRepo,
		txManager: txManager,
		eventTTL:  eventTTL,
	}
}
