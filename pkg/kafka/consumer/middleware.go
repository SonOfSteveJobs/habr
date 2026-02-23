package consumer

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

// Recovery - ловит панику, сообщение не будет помечено как прочитанное
func Recovery(next kafka.MessageHandler) kafka.MessageHandler {
	return func(ctx context.Context, msg kafka.Message) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log := logger.Ctx(ctx)
				log.Error().
					Str("topic", msg.Topic).
					Int32("partition", msg.Partition).
					Int64("offset", msg.Offset).
					Str("stack", string(debug.Stack())).
					Msgf("kafka: panic recovered: %v", r)

				err = fmt.Errorf("kafka: panic: %v", r)
			}
		}()

		return next(ctx, msg)
	}
}

// Logging - logger middleware
func Logging(next kafka.MessageHandler) kafka.MessageHandler {
	return func(ctx context.Context, msg kafka.Message) error {
		start := time.Now()

		err := next(ctx, msg)

		log := logger.Ctx(ctx)
		log.Debug().
			Err(err).
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Dur("duration", time.Since(start)).
			Msg("kafka: message processed")

		return err
	}
}

// WithRetry - ретраит сообщение несколько раз, если безуспешно - логирует и скипает.Можно добавить отправку в DLQ
func WithRetry(maxRetries int) Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			var err error

			for attempt := 1; attempt <= maxRetries; attempt++ {
				err = next(ctx, msg)
				if err == nil {
					return nil
				}

				log := logger.Ctx(ctx)
				log.Warn().
					Err(err).
					Int("attempt", attempt).
					Int("max_retries", maxRetries).
					Str("topic", msg.Topic).
					Int64("offset", msg.Offset).
					Msg("kafka: handler failed, retrying")
			}

			log := logger.Ctx(ctx)
			log.Error().
				Err(err).
				Str("topic", msg.Topic).
				Int32("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("kafka: message skipped after max retries")

			return nil
		}
	}
}
