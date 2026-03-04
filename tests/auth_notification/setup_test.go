//go:build integration

package auth_notification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/pkg/transaction"
)

// --- TestMain ---

func TestMain(m *testing.M) {
	logger.Init("error", false)
	os.Exit(m.Run())
}

// --- Event types (JSON contract between auth and notification) ---

type registeredEvent struct {
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type outboxEvent struct {
	EventID   uuid.UUID
	Topic     string
	Key       []byte
	Value     []byte
	CreatedAt time.Time
}

// --- Mock email sender ---

type mockEmailSender struct {
	mu       sync.Mutex
	sent     []registeredEvent
	attempts int
	failFn   func(registeredEvent) error
}

func newMockEmailSender() *mockEmailSender {
	return &mockEmailSender{}
}

func (m *mockEmailSender) Send(_ context.Context, event registeredEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts++
	if m.failFn != nil {
		if err := m.failFn(event); err != nil {
			return err
		}
	}
	m.sent = append(m.sent, event)
	return nil
}

func (m *mockEmailSender) setFailFn(fn func(registeredEvent) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failFn = fn
}

func (m *mockEmailSender) sentCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sent)
}

func (m *mockEmailSender) attemptCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.attempts
}

func (m *mockEmailSender) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = nil
	m.attempts = 0
	m.failFn = nil
}

// --- Controllable producer (for relay tests) ---

type controllableProducer struct {
	mu   sync.Mutex
	fail bool
	sent []kafka.Message
}

func (p *controllableProducer) Send(_ context.Context, msg kafka.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return fmt.Errorf("producer: forced failure")
	}
	p.sent = append(p.sent, msg)
	return nil
}

func (p *controllableProducer) setFail(fail bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fail = fail
}

func (p *controllableProducer) sentCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.sent)
}

func (p *controllableProducer) sentMessages() []kafka.Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]kafka.Message, len(p.sent))
	copy(result, p.sent)
	return result
}

func (p *controllableProducer) reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = nil
}

// --- Auth DB helpers ---

type authDB struct {
	pool *pgxpool.Pool
}

func (db *authDB) insertOutboxEvent(ctx context.Context, e outboxEvent) error {
	const query = `INSERT INTO outbox (event_id, topic, key, value, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.pool.Exec(ctx, query, e.EventID, e.Topic, e.Key, e.Value, e.CreatedAt)
	return err
}

func (db *authDB) fetchUnsentEvents(ctx context.Context, limit int) ([]outboxEvent, error) {
	const query = `SELECT event_id, topic, key, value, created_at FROM outbox WHERE NOT is_sent ORDER BY created_at LIMIT $1`
	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []outboxEvent
	for rows.Next() {
		var e outboxEvent
		if err := rows.Scan(&e.EventID, &e.Topic, &e.Key, &e.Value, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (db *authDB) markSent(ctx context.Context, eventID uuid.UUID) error {
	_, err := db.pool.Exec(ctx, `UPDATE outbox SET is_sent = TRUE WHERE event_id = $1`, eventID)
	return err
}

func (db *authDB) countUnsent(ctx context.Context) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE NOT is_sent`).Scan(&count)
	return count, err
}

func (db *authDB) truncate(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, `TRUNCATE outbox, users CASCADE`)
	return err
}

// --- Notification DB helpers ---

type notificationDB struct {
	pool *pgxpool.Pool
}

func (db *notificationDB) isProcessed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	var count int
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM processed_events WHERE event_id = $1`, eventID).Scan(&count)
	return count > 0, err
}

func (db *notificationDB) countProcessed(ctx context.Context) (int, error) {
	var count int
	err := db.pool.QueryRow(ctx, `SELECT count(*) FROM processed_events`).Scan(&count)
	return count, err
}

func (db *notificationDB) truncate(ctx context.Context) error {
	_, err := db.pool.Exec(ctx, `TRUNCATE processed_events CASCADE`)
	return err
}

// --- Notification handler (reproduces the real HandleEvent logic) ---

func newNotificationHandler(
	pool *pgxpool.Pool,
	sender *mockEmailSender,
	eventTTL time.Duration,
) kafka.MessageHandler {
	txm := transaction.New(pool)
	return func(ctx context.Context, msg kafka.Message) error {
		var event registeredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("unmarshal event: %w", err)
		}

		if time.Since(event.CreatedAt) > eventTTL {
			return nil
		}

		eventID, err := uuid.Parse(event.EventID)
		if err != nil {
			return fmt.Errorf("parse event ID: %w", err)
		}

		return txm.Wrap(ctx, func(ctx context.Context) error {
			const q = `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`
			ct, err := txm.ExtractExecutor(ctx).Exec(ctx, q, eventID)
			if err != nil {
				return fmt.Errorf("mark processed: %w", err)
			}

			if ct.RowsAffected() == 0 {
				return nil
			}

			if err := sender.Send(ctx, event); err != nil {
				return fmt.Errorf("send email: %w", err)
			}

			return nil
		})
	}
}

// --- Relay poll (reproduces the real relay poll logic) ---

type relayProducer interface {
	Send(ctx context.Context, msg kafka.Message) error
}

func relayPoll(ctx context.Context, pool *pgxpool.Pool, prod relayProducer, limit int) (int, error) {
	const query = `SELECT event_id, topic, key, value, created_at FROM outbox WHERE NOT is_sent ORDER BY created_at LIMIT $1`

	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return 0, fmt.Errorf("fetch unsent: %w", err)
	}
	defer rows.Close()

	var events []outboxEvent
	for rows.Next() {
		var e outboxEvent
		if err := rows.Scan(&e.EventID, &e.Topic, &e.Key, &e.Value, &e.CreatedAt); err != nil {
			return 0, fmt.Errorf("scan outbox: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("rows error: %w", err)
	}

	var sent int
	for _, event := range events {
		msg := kafka.Message{
			Key:      event.Key,
			Value:    event.Value,
			Metadata: event.EventID.String(),
		}
		if err := prod.Send(ctx, msg); err != nil {
			return sent, fmt.Errorf("send event %s: %w", event.EventID, err)
		}
		sent++
	}

	return sent, nil
}

// --- Helpers ---

func buildEventJSON(t testing.TB, eventID, userID, email, code string, createdAt time.Time) []byte {
	t.Helper()
	event := registeredEvent{
		EventID:   eventID,
		UserID:    userID,
		Email:     email,
		Code:      code,
		CreatedAt: createdAt,
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	return data
}

func makeOutboxEvent(t testing.TB, topic string) (outboxEvent, registeredEvent) {
	t.Helper()
	eventID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	reg := registeredEvent{
		EventID:   eventID.String(),
		UserID:    userID.String(),
		Email:     fmt.Sprintf("user-%s@test.com", userID.String()[:8]),
		Code:      "123456",
		CreatedAt: now,
	}

	value, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	out := outboxEvent{
		EventID:   eventID,
		Topic:     topic,
		Key:       []byte(userID.String()),
		Value:     value,
		CreatedAt: now,
	}

	return out, reg
}

func createTestTopic(t testing.TB, brokers []string, topic string) {
	t.Helper()
	cfg := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		t.Fatalf("create kafka admin: %v", err)
	}
	defer admin.Close()

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		t.Fatalf("create topic %s: %v", topic, err)
	}
}

func waitFor(t testing.TB, timeout, interval time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("timeout waiting for: %s", msg)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}
