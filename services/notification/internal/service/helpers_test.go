package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/model"
)

const testEventTTL = 15 * time.Minute

type mockEventRepo struct {
	markProcessedFn     func(ctx context.Context, eventID uuid.UUID) (bool, error)
	markProcessedCalled bool
}

func (m *mockEventRepo) MarkProcessed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	m.markProcessedCalled = true
	return m.markProcessedFn(ctx, eventID)
}

type mockEmailSender struct {
	sendFn     func(ctx context.Context, event model.UserRegisteredEvent) error
	sendCalled bool
}

func (m *mockEmailSender) Send(ctx context.Context, event model.UserRegisteredEvent) error {
	m.sendCalled = true
	return m.sendFn(ctx, event)
}

type mockTxManager struct{}

func (m *mockTxManager) Wrap(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

func newTestService(eventRepo *mockEventRepo, emailSender *mockEmailSender) *Service {
	return New(eventRepo, &mockTxManager{}, emailSender, testEventTTL)
}

func testEvent(t *testing.T) model.UserRegisteredEvent {
	t.Helper()

	return model.UserRegisteredEvent{
		EventID:   uuid.Must(uuid.NewV7()).String(),
		UserID:    uuid.Must(uuid.NewV7()).String(),
		Email:     "user@example.com",
		Code:      "123456",
		CreatedAt: time.Now(),
	}
}

func testMessage(t *testing.T, event model.UserRegisteredEvent) kafka.Message {
	t.Helper()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	return kafka.Message{Value: data}
}
