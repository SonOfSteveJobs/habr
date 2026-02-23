package model

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	EventID   uuid.UUID
	Topic     string
	Key       []byte
	Value     []byte
	CreatedAt time.Time
}
