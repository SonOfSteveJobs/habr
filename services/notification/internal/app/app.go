package app

import (
	"context"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

type App struct {
	infra   *infraContainer
	service *serviceContainer
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() {
	log := logger.Logger()

	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	closer.AddNamed("kafka consumer", func(_ context.Context) error {
		consumerCancel()
		return nil
	})

	go func() {
		handler := func(ctx context.Context, msg kafka.Message) error {
			log.Info().
				Str("topic", msg.Topic).
				Int32("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("received message")
			return nil
		}

		if err := a.service.KafkaConsumer().Consume(consumerCtx, handler); err != nil {
			log.Error().Err(err).Msg("kafka consumer failed")
		}
	}()

	log.Info().Msg("notification service started")

	closer.Wait()
}

type initStep struct {
	name string
	fn   func(context.Context) error
}

func (a *App) initDeps(ctx context.Context) error {
	log := logger.Logger()

	steps := []initStep{
		{"infra: postgres, kafka", a.initInfra},
		{"service: consumer", a.initService},
	}

	for _, s := range steps {
		if err := s.fn(ctx); err != nil {
			log.Error().Err(err).Str("component", s.name).Msg("init failed")
			return err
		}
		log.Info().Str("component", s.name).Msg("init ok")
	}

	return nil
}

func (a *App) initInfra(ctx context.Context) error {
	infra, err := newInfraContainer(ctx)
	if err != nil {
		return err
	}
	a.infra = infra
	return nil
}

func (a *App) initService(_ context.Context) error {
	a.service = newServiceContainer(a.infra)
	return nil
}
