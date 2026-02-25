package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/model"
)

func (s *Service) HandleEvent(ctx context.Context, msg kafka.Message) error {
	log := logger.Logger()

	var event model.UserRegisteredEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal event")
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if time.Since(event.CreatedAt) > s.eventTTL {
		log.Warn().
			Str("event_id", event.EventID).
			Time("created_at", event.CreatedAt).
			Msg("event TTL expired, skipping")
		return nil
	}

	eventID, err := uuid.Parse(event.EventID)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.EventID).Msg("invalid event ID")
		return fmt.Errorf("parse event ID: %w", err)
	}

	return s.txManager.Wrap(ctx, func(ctx context.Context) error {
		inserted, err := s.eventRepo.MarkProcessed(ctx, eventID)
		if err != nil {
			return fmt.Errorf("mark processed: %w", err)
		}

		if !inserted {
			log.Info().Str("event_id", event.EventID).Msg("duplicate event, skipping")
			return nil
		}

		if err := s.emailSender.Send(ctx, event); err != nil {
			return fmt.Errorf("send email: %w", err)
		}

		return nil
	})
}
