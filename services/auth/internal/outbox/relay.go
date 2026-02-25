package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type OutboxRepository interface {
	FetchUnsent(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkSent(ctx context.Context, eventID uuid.UUID) error
	DeleteSent(ctx context.Context) error
}

type Producer interface {
	Send(ctx context.Context, msg kafka.Message) error
}

type Relay struct {
	repo            OutboxRepository
	producer        Producer
	pollInterval    time.Duration
	cleanupInterval time.Duration
	fetchLimit      int
}

func NewRelay(repo OutboxRepository, producer Producer, pollInterval, cleanupInterval time.Duration, fetchLimit int) *Relay {
	return &Relay{
		repo:            repo,
		producer:        producer,
		pollInterval:    pollInterval,
		cleanupInterval: cleanupInterval,
		fetchLimit:      fetchLimit,
	}
}

func (r *Relay) Run(ctx context.Context) {
	log := logger.Logger()

	pollTicker := time.NewTicker(r.pollInterval)
	defer pollTicker.Stop()

	cleanupTicker := time.NewTicker(r.cleanupInterval)
	defer cleanupTicker.Stop()

	log.Info().Msg("outbox relay started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("outbox relay stopped")
			return

		case <-pollTicker.C:
			r.poll(ctx)

		case <-cleanupTicker.C:
			r.cleanup(ctx)
		}
	}
}

func (r *Relay) poll(ctx context.Context) {
	log := logger.Logger()

	events, err := r.repo.FetchUnsent(ctx, r.fetchLimit)
	if err != nil {
		log.Error().Err(err).Msg("outbox: fetch unsent failed")
		return
	}

	for _, event := range events {
		msg := kafka.Message{
			Key:      event.Key,
			Value:    event.Value,
			Metadata: event.EventID.String(),
		}

		log.Info().Str("event_id", event.EventID.String()).Msg("outbox: send to kafka")

		if err := r.producer.Send(ctx, msg); err != nil {
			log.Error().Err(err).
				Str("event_id", event.EventID.String()).
				Msg("outbox: send to kafka failed")
			return
		}
	}
}

func (r *Relay) cleanup(ctx context.Context) {
	log := logger.Logger()
	log.Info().Msg("outbox: cleanup")

	if err := r.repo.DeleteSent(ctx); err != nil {
		log.Error().Err(err).Msg("outbox: cleanup failed")
	}
}
