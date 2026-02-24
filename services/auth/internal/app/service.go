package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/producer"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
	authgrpc "github.com/SonOfSteveJobs/habr/services/auth/internal/handler/grpc"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/outbox"
	outboxrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/outbox"
	tokenrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/token"
	userrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/user"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/service"
)

type serviceContainer struct {
	infra *infraContainer

	outboxRepo    *outboxrepo.Repository
	userRepo      *userrepo.Repository
	tokenRepo     *tokenrepo.Repository
	kafkaProducer *producer.AsyncProducer
	outboxRelay   *outbox.Relay
	authService   *service.Service
	handler       *authgrpc.Handler
}

func newServiceContainer(infra *infraContainer) *serviceContainer {
	return &serviceContainer{infra: infra}
}

func (c *serviceContainer) OutboxRepo() *outboxrepo.Repository {
	if c.outboxRepo == nil {
		c.outboxRepo = outboxrepo.New(c.infra.TxManager())
	}

	return c.outboxRepo
}

func (c *serviceContainer) UserRepo() *userrepo.Repository {
	if c.userRepo == nil {
		c.userRepo = userrepo.New(c.infra.TxManager())
	}

	return c.userRepo
}

func (c *serviceContainer) TokenRepo() *tokenrepo.Repository {
	if c.tokenRepo == nil {
		c.tokenRepo = tokenrepo.New(c.infra.RedisClient())
	}

	return c.tokenRepo
}

func (c *serviceContainer) KafkaProducer() *producer.AsyncProducer {
	if c.kafkaProducer == nil {
		log := logger.Logger()
		repo := c.OutboxRepo()

		onSuccess := func(metadata any) {
			eventID, ok := metadata.(string)
			if !ok {
				return
			}

			uid, err := uuid.Parse(eventID)
			if err != nil {
				log.Error().Err(err).Str("event_id", eventID).Msg("outbox: invalid event_id in metadata")
				return
			}

			if err := repo.MarkSent(context.Background(), uid); err != nil {
				log.Error().Err(err).Str("event_id", eventID).Msg("outbox: mark sent failed")
			}
		}

		p := producer.NewAsync(c.infra.SaramaProducer(), config.AppConfig().Kafka().Topic(), onSuccess)
		closer.AddNamed("kafka producer", func(_ context.Context) error {
			return p.Close()
		})

		c.kafkaProducer = p
	}

	return c.kafkaProducer
}

func (c *serviceContainer) OutboxRelay() *outbox.Relay {
	if c.outboxRelay == nil {
		cfg := config.AppConfig().Kafka()
		c.outboxRelay = outbox.NewRelay(
			c.OutboxRepo(),
			c.KafkaProducer(),
			cfg.OutboxPollInterval(),
			cfg.OutboxCleanupInterval(),
			cfg.OutboxFetchLimit(),
		)
	}

	return c.outboxRelay
}

func (c *serviceContainer) AuthService() *service.Service {
	if c.authService == nil {
		cfg := config.AppConfig()
		c.authService = service.New(
			c.UserRepo(), c.TokenRepo(), c.OutboxRepo(), c.infra.TxManager(),
			cfg.JWTSecret(), cfg.Kafka().Topic(),
			cfg.AccessTokenTTL(), cfg.RefreshTokenTTL(),
		)
	}

	return c.authService
}

func (c *serviceContainer) Handler() *authgrpc.Handler {
	if c.handler == nil {
		c.handler = authgrpc.New(c.AuthService())
	}

	return c.handler
}
