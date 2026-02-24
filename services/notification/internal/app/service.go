package app

import (
	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/config"
)

type serviceContainer struct {
	infra *infraContainer

	kafkaConsumer *consumer.Consumer
}

func newServiceContainer(infra *infraContainer) *serviceContainer {
	return &serviceContainer{infra: infra}
}

func (c *serviceContainer) KafkaConsumer() *consumer.Consumer {
	if c.kafkaConsumer == nil {
		cfg := config.AppConfig().Kafka()
		c.kafkaConsumer = consumer.New(
			c.infra.ConsumerGroup(),
			[]string{cfg.Topic()},
			consumer.Recovery,
			consumer.Logging,
			consumer.WithRetry(3),
		)
	}

	return c.kafkaConsumer
}
