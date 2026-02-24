package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/model"
)

func TestHandleEvent_Success(t *testing.T) {
	event := testEvent(t)

	eventRepo := &mockEventRepo{
		markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
	}
	emailSender := &mockEmailSender{
		sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error { return nil },
	}
	svc := newTestService(eventRepo, emailSender)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !eventRepo.markProcessedCalled {
		t.Error("eventRepo.MarkProcessed was not called")
	}

	if !emailSender.sendCalled {
		t.Error("emailSender.Send was not called")
	}
}

func TestHandleEvent_DuplicateEvent(t *testing.T) {
	event := testEvent(t)

	eventRepo := &mockEventRepo{
		markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil },
	}
	emailSender := &mockEmailSender{
		sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error {
			t.Error("emailSender.Send should not be called for duplicate event")
			return nil
		},
	}
	svc := newTestService(eventRepo, emailSender)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !eventRepo.markProcessedCalled {
		t.Error("eventRepo.MarkProcessed was not called")
	}

	if emailSender.sendCalled {
		t.Error("emailSender.Send should not be called for duplicate event")
	}
}

func TestHandleEvent_ExpiredTTL(t *testing.T) {
	event := testEvent(t)
	event.CreatedAt = time.Now().Add(-2 * testEventTTL)

	eventRepo := &mockEventRepo{
		markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) {
			t.Error("eventRepo.MarkProcessed should not be called for expired event")
			return false, nil
		},
	}
	emailSender := &mockEmailSender{
		sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error {
			t.Error("emailSender.Send should not be called for expired event")
			return nil
		},
	}
	svc := newTestService(eventRepo, emailSender)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleEvent_InvalidJSON(t *testing.T) {
	svc := newTestService(
		&mockEventRepo{markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil }},
		&mockEmailSender{sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error { return nil }},
	)

	msg := kafka.Message{Value: []byte("not-json")}

	err := svc.HandleEvent(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleEvent_InvalidEventID(t *testing.T) {
	event := testEvent(t)
	event.EventID = "not-a-uuid"

	svc := newTestService(
		&mockEventRepo{markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, nil }},
		&mockEmailSender{sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error { return nil }},
	)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if err == nil {
		t.Fatal("expected error for invalid event ID")
	}
}

func TestHandleEvent_MarkProcessedError(t *testing.T) {
	event := testEvent(t)
	repoErr := errors.New("db connection refused")

	eventRepo := &mockEventRepo{
		markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return false, repoErr },
	}
	emailSender := &mockEmailSender{
		sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error { return nil },
	}
	svc := newTestService(eventRepo, emailSender)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if !errors.Is(err, repoErr) {
		t.Errorf("error = %v, want %v", err, repoErr)
	}

	if emailSender.sendCalled {
		t.Error("emailSender.Send should not be called when MarkProcessed fails")
	}
}

func TestHandleEvent_EmailSenderError(t *testing.T) {
	event := testEvent(t)
	sendErr := errors.New("smtp connection refused")

	eventRepo := &mockEventRepo{
		markProcessedFn: func(_ context.Context, _ uuid.UUID) (bool, error) { return true, nil },
	}
	emailSender := &mockEmailSender{
		sendFn: func(_ context.Context, _ model.UserRegisteredEvent) error { return sendErr },
	}
	svc := newTestService(eventRepo, emailSender)

	err := svc.HandleEvent(context.Background(), testMessage(t, event))
	if !errors.Is(err, sendErr) {
		t.Errorf("error = %v, want %v", err, sendErr)
	}
}
