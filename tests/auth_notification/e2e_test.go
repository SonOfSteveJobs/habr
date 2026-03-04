//go:build integration

package auth_notification

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/producer"
	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

func TestE2E(t *testing.T) {
	root := testinfra.ProjectRoot(t)
	authPG := testinfra.NewPostgres(t, filepath.Join(root, "migrations/auth"))
	notifPG := testinfra.NewPostgres(t, filepath.Join(root, "migrations/notification"))
	kf := testinfra.NewKafka(t)

	adb := &authDB{pool: authPG.Pool}
	ndb := &notificationDB{pool: notifPG.Pool}
	ctx := context.Background()

	t.Run("notification_down_during_registration", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		require.NoError(t, ndb.truncate(ctx))

		topic := fmt.Sprintf("test-notif-down-%s", uuid.New().String()[:8])
		createTestTopic(t, kf.Brokers, topic)

		sender := newMockEmailSender()

		// 1. Auth: регистрация 3 пользователей → события в outbox
		var events []outboxEvent
		for i := 0; i < 3; i++ {
			out, _ := makeOutboxEvent(t, topic)
			require.NoError(t, adb.insertOutboxEvent(ctx, out))
			events = append(events, out)
		}

		// 2. Relay: отправляет события в Kafka через sync producer
		syncProd := newSyncProducer(t, kf.Brokers, topic)
		sent, err := relayPoll(ctx, authPG.Pool, syncProd, 100)
		require.NoError(t, err)
		assert.Equal(t, 3, sent)

		// 3. Notification "лежит" — события сидят в Kafka
		//    Пауза для имитации downtime
		time.Sleep(2 * time.Second)

		// 4. Notification поднимается — запускаем consumer
		handler := newNotificationHandler(notifPG.Pool, sender, eventTTL)
		groupID := fmt.Sprintf("test-group-%s", uuid.New().String()[:8])
		kafkaConsumer := newKafkaConsumer(t, kf.Brokers, groupID, topic)

		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			_ = kafkaConsumer.Consume(consumerCtx, handler)
		}()

		// 5. Ждём обработки всех 3 событий
		waitFor(t, 30*time.Second, 500*time.Millisecond, func() bool {
			count, _ := ndb.countProcessed(ctx)
			return count == 3
		}, "3 events processed after notification recovery")

		assert.Equal(t, 3, sender.sentCount(), "all 3 emails should be sent")
		cancel()
	})

	t.Run("duplicate_delivery_idempotent", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		require.NoError(t, ndb.truncate(ctx))

		topic := fmt.Sprintf("test-dedup-%s", uuid.New().String()[:8])
		createTestTopic(t, kf.Brokers, topic)

		sender := newMockEmailSender()

		// создаём одно событие
		out, _ := makeOutboxEvent(t, topic)
		require.NoError(t, adb.insertOutboxEvent(ctx, out))

		// relay отправляет в Kafka
		syncProd := newSyncProducer(t, kf.Brokers, topic)
		sent, err := relayPoll(ctx, authPG.Pool, syncProd, 100)
		require.NoError(t, err)
		assert.Equal(t, 1, sent)

		// "рестарт relay" — событие не помечено sent, relay отправит повторно
		sent, err = relayPoll(ctx, authPG.Pool, syncProd, 100)
		require.NoError(t, err)
		assert.Equal(t, 1, sent, "relay sends duplicate after restart")

		// Kafka теперь содержит 2 сообщения с одним event_id
		// Notification consumer обрабатывает оба — второе дедуплицируется
		handler := newNotificationHandler(notifPG.Pool, sender, eventTTL)
		groupID := fmt.Sprintf("test-group-%s", uuid.New().String()[:8])
		kafkaConsumer := newKafkaConsumer(t, kf.Brokers, groupID, topic)

		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			_ = kafkaConsumer.Consume(consumerCtx, handler)
		}()

		// ждём обработки — должен быть 1 email, не 2
		waitFor(t, 30*time.Second, 500*time.Millisecond, func() bool {
			return sender.attemptCount() >= 1
		}, "at least 1 email attempt")

		// даём время на обработку второго сообщения
		time.Sleep(3 * time.Second)

		assert.Equal(t, 1, sender.sentCount(), "only 1 email despite 2 kafka messages")

		count, err := ndb.countProcessed(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "only 1 record in processed_events")
		cancel()
	})

	t.Run("partial_email_failure_with_retry", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		require.NoError(t, ndb.truncate(ctx))

		topic := fmt.Sprintf("test-partial-%s", uuid.New().String()[:8])
		createTestTopic(t, kf.Brokers, topic)

		// создаём 3 события
		var outEvents []outboxEvent
		var regEvents []registeredEvent
		for i := 0; i < 3; i++ {
			out, reg := makeOutboxEvent(t, topic)
			require.NoError(t, adb.insertOutboxEvent(ctx, out))
			outEvents = append(outEvents, out)
			regEvents = append(regEvents, reg)
		}

		// event[1] будет фейлить email
		failEventID := regEvents[1].EventID
		sender := newMockEmailSender()
		sender.setFailFn(func(event registeredEvent) error {
			if event.EventID == failEventID {
				return fmt.Errorf("smtp: rejected for %s", failEventID)
			}
			return nil
		})

		// relay отправляет все 3 в Kafka
		syncProd := newSyncProducer(t, kf.Brokers, topic)
		sent, err := relayPoll(ctx, authPG.Pool, syncProd, 100)
		require.NoError(t, err)
		assert.Equal(t, 3, sent)

		// consumer с retry middleware (3 попытки)
		handler := newNotificationHandler(notifPG.Pool, sender, eventTTL)
		groupID := fmt.Sprintf("test-group-%s", uuid.New().String()[:8])
		kafkaConsumer := newKafkaConsumerWithRetry(t, kf.Brokers, groupID, topic, 3)

		consumerCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() {
			_ = kafkaConsumer.Consume(consumerCtx, handler)
		}()

		// ждём обработки успешных событий
		waitFor(t, 30*time.Second, 500*time.Millisecond, func() bool {
			return sender.sentCount() >= 2
		}, "at least 2 emails sent")

		// даём время на retry event[1] и обработку оставшихся
		time.Sleep(5 * time.Second)

		// 2 события успешно обработаны, 1 зафейлено
		assert.Equal(t, 2, sender.sentCount(), "2 emails sent successfully")

		// event[1] НЕ в processed_events (транзакция откатилась при каждом retry,
		// а после max retries WithRetry скипнул сообщение)
		processed, err := ndb.isProcessed(ctx, uuid.MustParse(failEventID))
		require.NoError(t, err)
		assert.False(t, processed, "failed event should NOT be in processed_events")

		// 2 записи в processed_events
		count, err := ndb.countProcessed(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "2 events in processed_events")

		cancel()
	})
}

// --- E2E helpers ---

func newSyncProducer(t testing.TB, brokers []string, topic string) *producer.SyncProducer {
	t.Helper()
	cfg := producer.NewSyncConfig()
	sp, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		t.Fatalf("create sync producer: %v", err)
	}
	t.Cleanup(func() { _ = sp.Close() })
	return producer.NewSync(sp, topic)
}

func newKafkaConsumer(t testing.TB, brokers []string, groupID, topic string) *consumer.Consumer {
	t.Helper()
	return newKafkaConsumerWithRetry(t, brokers, groupID, topic, 0)
}

func newKafkaConsumerWithRetry(t testing.TB, brokers []string, groupID, topic string, maxRetries int) *consumer.Consumer {
	t.Helper()
	cfg := consumer.NewConfig()
	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		t.Fatalf("create consumer group: %v", err)
	}
	t.Cleanup(func() { _ = group.Close() })

	middlewares := []consumer.Middleware{consumer.Recovery, consumer.Logging}
	if maxRetries > 0 {
		middlewares = append(middlewares, consumer.WithRetry(maxRetries))
	}

	return consumer.New(group, []string{topic}, middlewares...)
}
