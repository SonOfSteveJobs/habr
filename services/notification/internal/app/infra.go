package app

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/config"
)

type infraContainer struct {
	txManager     *transaction.Manager
	consumerGroup sarama.ConsumerGroup
}

func newInfraContainer(ctx context.Context) (*infraContainer, error) {
	c := &infraContainer{}

	if err := c.initPgPool(ctx); err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	if err := c.initConsumerGroup(); err != nil {
		return nil, fmt.Errorf("kafka consumer group: %w", err)
	}

	return c, nil
}

func (c *infraContainer) TxManager() *transaction.Manager     { return c.txManager }
func (c *infraContainer) ConsumerGroup() sarama.ConsumerGroup { return c.consumerGroup }

func (c *infraContainer) initPgPool(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, config.AppConfig().DBURI())
	if err != nil {
		return err
	}
	closer.AddNamed("postgres", func(_ context.Context) error {
		pool.Close()
		return nil
	})

	c.txManager = transaction.New(pool)
	return nil
}

func (c *infraContainer) initConsumerGroup() error {
	cfg := config.AppConfig().Kafka()

	saramaCfg := consumer.NewConfig()
	group, err := sarama.NewConsumerGroup(cfg.Brokers(), cfg.GroupID(), saramaCfg)
	if err != nil {
		return err
	}
	closer.AddNamed("kafka consumer group", func(_ context.Context) error {
		return group.Close()
	})

	c.consumerGroup = group
	return nil
}
