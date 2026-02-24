package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

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

	err = s.txManager.Wrap(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		outboxEvent, err := s.buildOutboxEvent(user)
		if err != nil {
			return err
		}

		return s.outboxRepo.Insert(ctx, outboxEvent)
	})
	if err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (s *Service) buildOutboxEvent(user *model.User) (model.OutboxEvent, error) {
	event := UserRegisteredEvent{
		EventID:   uuid.New().String(),
		UserID:    user.ID.String(),
		Email:     user.Email,
		CreatedAt: time.Now(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		return model.OutboxEvent{}, fmt.Errorf("marshal event: %w", err)
	}

	return model.OutboxEvent{
		EventID:   uuid.MustParse(event.EventID),
		Topic:     s.kafkaTopic,
		Key:       []byte(user.ID.String()),
		Value:     value,
		CreatedAt: event.CreatedAt,
	}, nil
}
