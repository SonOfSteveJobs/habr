package app

import (
	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/config"
	eventRepo "github.com/SonOfSteveJobs/habr/services/notification/internal/repository/event"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/service"
)

type serviceContainer struct {
	infra *infraContainer

	kafkaConsumer       *consumer.Consumer
	eventRepository     *eventRepo.Repository
	notificationService *service.Service
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

func (c *serviceContainer) EventRepository() *eventRepo.Repository {
	if c.eventRepository == nil {
		c.eventRepository = eventRepo.New(c.infra.TxManager())
	}

	return c.eventRepository
}

func (c *serviceContainer) NotificationService() *service.Service {
	if c.notificationService == nil {
		c.notificationService = service.New(
			c.EventRepository(),
			c.infra.TxManager(),
			config.AppConfig().EventTTL(),
		)
	}

	return c.notificationService
}
