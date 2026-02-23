package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type UserRegisteredEvent struct {
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Service) Register(ctx context.Context, email, password string) (uuid.UUID, error) {
	user, err := model.NewUser(email, password)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return uuid.Nil, err
	}

	if err := s.publishUserRegistered(ctx, user); err != nil {
		log := logger.Ctx(ctx)
		log.Error().Err(err).Msg("failed to publish user-registered event")
	}

	return user.ID, nil
}

func (s *Service) publishUserRegistered(ctx context.Context, user *model.User) error {
	event := UserRegisteredEvent{
		EventID:   uuid.New().String(),
		UserID:    user.ID.String(),
		Email:     user.Email,
		CreatedAt: time.Now(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:      []byte(user.ID.String()),
		Value:    value,
		Metadata: event.EventID,
	}

	return s.producer.Send(ctx, msg)
}
