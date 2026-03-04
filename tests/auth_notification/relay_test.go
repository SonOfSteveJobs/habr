//go:build integration

package auth_notification

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

const testTopic = "user-registered"

func TestRelay(t *testing.T) {
	pg := testinfra.NewPostgres(t, filepath.Join(testinfra.ProjectRoot(t), "migrations/auth"))
	adb := &authDB{pool: pg.Pool}
	ctx := context.Background()

	t.Run("producer_failure_events_stay_in_outbox", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		prod := &controllableProducer{fail: true}

		// вставляем 3 события в outbox
		for i := 0; i < 3; i++ {
			out, _ := makeOutboxEvent(t, testTopic)
			require.NoError(t, adb.insertOutboxEvent(ctx, out))
		}

		// poll с падающим producer — первое событие фейлит, relay останавливается
		sent, err := relayPoll(ctx, pg.Pool, prod, 100)
		require.Error(t, err)
		assert.Equal(t, 0, sent)
		assert.Equal(t, 0, prod.sentCount())

		// все события остались в outbox
		unsent, err := adb.countUnsent(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, unsent, "all events should remain unsent")
	})

	t.Run("producer_recovery_events_delivered", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		prod := &controllableProducer{fail: true}

		// вставляем события
		for i := 0; i < 3; i++ {
			out, _ := makeOutboxEvent(t, testTopic)
			require.NoError(t, adb.insertOutboxEvent(ctx, out))
		}

		// первый poll — producer падает
		sent, err := relayPoll(ctx, pg.Pool, prod, 100)
		require.Error(t, err)
		assert.Equal(t, 0, sent)

		// "чиним" producer
		prod.setFail(false)

		// второй poll — события доставлены
		sent, err = relayPoll(ctx, pg.Pool, prod, 100)
		require.NoError(t, err)
		assert.Equal(t, 3, sent)
		assert.Equal(t, 3, prod.sentCount())

		// события всё ещё unsent в БД (MarkSent вызывается async callback-ом, не relay)
		unsent, err := adb.countUnsent(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, unsent, "events still unsent until async callback marks them")
	})

	t.Run("mark_sent_after_delivery", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		prod := &controllableProducer{}

		// вставляем событие
		out, _ := makeOutboxEvent(t, testTopic)
		require.NoError(t, adb.insertOutboxEvent(ctx, out))

		// relay отправляет
		sent, err := relayPoll(ctx, pg.Pool, prod, 100)
		require.NoError(t, err)
		assert.Equal(t, 1, sent)

		// симулируем async callback — markSent
		require.NoError(t, adb.markSent(ctx, out.EventID))

		// событие больше не unsent
		unsent, err := adb.countUnsent(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, unsent)

		// повторный poll — ничего не найдено
		prod.reset()
		sent, err = relayPoll(ctx, pg.Pool, prod, 100)
		require.NoError(t, err)
		assert.Equal(t, 0, sent)
	})

	t.Run("partial_failure_stops_at_first_error", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))

		// вставляем 5 событий с разным created_at для детерминированного ORDER BY
		for i := 0; i < 5; i++ {
			out, _ := makeOutboxEvent(t, testTopic)
			time.Sleep(time.Millisecond)
			require.NoError(t, adb.insertOutboxEvent(ctx, out))
		}

		// producer который фейлит на 3-м сообщении
		failAfter := &failAfterNProducer{maxSuccess: 2}

		sent, err := relayPoll(ctx, pg.Pool, failAfter, 100)
		require.Error(t, err)
		assert.Equal(t, 2, sent, "should send 2 events before failure")

		// все 5 событий unsent (relay не делает markSent)
		unsent, err := adb.countUnsent(ctx)
		require.NoError(t, err)
		assert.Equal(t, 5, unsent)
	})

	t.Run("empty_outbox_noop", func(t *testing.T) {
		require.NoError(t, adb.truncate(ctx))
		prod := &controllableProducer{}

		sent, err := relayPoll(ctx, pg.Pool, prod, 100)
		require.NoError(t, err)
		assert.Equal(t, 0, sent)
		assert.Equal(t, 0, prod.sentCount())
	})
}

// failAfterNProducer fails after N successful sends.
type failAfterNProducer struct {
	mu         sync.Mutex
	count      int
	maxSuccess int
}

func (p *failAfterNProducer) Send(_ context.Context, _ kafka.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.count++
	if p.count > p.maxSuccess {
		return fmt.Errorf("producer: fail after %d", p.maxSuccess)
	}
	return nil
}
