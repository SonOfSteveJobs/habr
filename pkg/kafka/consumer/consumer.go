package consumer

import (
	"context"
	"errors"

	"github.com/IBM/sarama"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

type Consumer struct {
	group       sarama.ConsumerGroup
	topics      []string
	middlewares []Middleware
}

// New - создаем новую консьюмер группу
// порядок мидллвар: Recovery → Logging → WithRetry → handler.
func New(group sarama.ConsumerGroup, topics []string, middlewares ...Middleware) *Consumer {
	return &Consumer{
		group:       group,
		topics:      topics,
		middlewares: middlewares,
	}
}

// Consume - отвечает за старт consumer loop. Поддерживает ребалансировку
func (c *Consumer) Consume(ctx context.Context, handler kafka.MessageHandler) error {
	gh := newGroupHandler(handler, c.middlewares...)

	for {
		if err := c.group.Consume(ctx, c.topics, gh); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}

			log := logger.Ctx(ctx)
			log.Error().Err(err).Msg("kafka: consume error")

			continue
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		log := logger.Ctx(ctx)
		log.Info().Msg("kafka: consumer group rebalancing")
	}
}

func (c *Consumer) Close() error {
	return c.group.Close()
}

// NewConfig - дефольный конфиг для консьюмер группы
func NewConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	return config
}
