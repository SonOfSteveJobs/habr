//go:build integration

package auth_notification

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

const eventTTL = 15 * time.Minute

func TestConsumer(t *testing.T) {
	pg := testinfra.NewPostgres(t, filepath.Join(testinfra.ProjectRoot(t), "migrations/notification"))
	ndb := &notificationDB{pool: pg.Pool}
	ctx := context.Background()

	t.Run("duplicate_event_idempotent", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		eventID := uuid.New().String()
		eventJSON := buildEventJSON(t, eventID, uuid.New().String(), "dup@test.com", "111111", time.Now())
		msg := kafka.Message{Value: eventJSON}

		// первая обработка — email отправлен
		err := handler(ctx, msg)
		require.NoError(t, err)
		assert.Equal(t, 1, sender.sentCount())

		// повторная обработка того же event_id — дубль, email не отправлен
		err = handler(ctx, msg)
		require.NoError(t, err)
		assert.Equal(t, 1, sender.sentCount(), "email should not be sent for duplicate event")

		// в processed_events ровно одна запись
		count, err := ndb.countProcessed(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("expired_ttl_skipped", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		// событие создано 20 минут назад, TTL = 15 минут → просрочено
		eventJSON := buildEventJSON(t, uuid.New().String(), uuid.New().String(), "expired@test.com", "222222", time.Now().Add(-20*time.Minute))
		msg := kafka.Message{Value: eventJSON}

		err := handler(ctx, msg)
		require.NoError(t, err)
		assert.Equal(t, 0, sender.sentCount(), "expired event should be skipped")

		count, err := ndb.countProcessed(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "expired event should not be in processed_events")
	})

	t.Run("invalid_json_returns_error", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		msg := kafka.Message{Value: []byte("not valid json {")}

		err := handler(ctx, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
		assert.Equal(t, 0, sender.sentCount())
	})

	t.Run("email_sender_error_rollback", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		sender.setFailFn(func(_ registeredEvent) error {
			return fmt.Errorf("smtp: connection refused")
		})
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		eventID := uuid.New()
		eventJSON := buildEventJSON(t, eventID.String(), uuid.New().String(), "fail@test.com", "333333", time.Now())
		msg := kafka.Message{Value: eventJSON}

		// обработка фейлит — транзакция откатывается
		err := handler(ctx, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send email")

		// event НЕ в processed_events (rollback)
		processed, err := ndb.isProcessed(ctx, eventID)
		require.NoError(t, err)
		assert.False(t, processed, "event should NOT be in processed_events after rollback")

		// фиксим email sender, повторяем
		sender.setFailFn(nil)
		err = handler(ctx, msg)
		require.NoError(t, err)

		// теперь event в processed_events
		processed, err = ndb.isProcessed(ctx, eventID)
		require.NoError(t, err)
		assert.True(t, processed, "event should be in processed_events after retry")
		assert.Equal(t, 1, sender.sentCount())
	})

	t.Run("invalid_event_id_returns_error", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		event := registeredEvent{
			EventID:   "not-a-uuid",
			UserID:    uuid.New().String(),
			Email:     "bad-id@test.com",
			Code:      "444444",
			CreatedAt: time.Now(),
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		err = handler(ctx, kafka.Message{Value: data})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse event ID")
		assert.Equal(t, 0, sender.sentCount())
	})

	t.Run("missing_required_fields", func(t *testing.T) {
		require.NoError(t, ndb.truncate(ctx))
		sender := newMockEmailSender()
		handler := newNotificationHandler(pg.Pool, sender, eventTTL)

		// пустой event_id → uuid.Parse вернёт ошибку
		event := registeredEvent{
			EventID:   "",
			UserID:    uuid.New().String(),
			Email:     "empty@test.com",
			Code:      "555555",
			CreatedAt: time.Now(),
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		err = handler(ctx, kafka.Message{Value: data})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse event ID")
	})
}
