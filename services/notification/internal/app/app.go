package app

import (
	"context"
	"time"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
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
		if err := a.service.KafkaConsumer().Consume(consumerCtx, a.service.NotificationService().HandleEvent); err != nil {
			log.Error().Err(err).Msg("kafka consumer failed")
		}
	}()

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	closer.AddNamed("cleanup", func(_ context.Context) error {
		cleanupCancel()
		return nil
	})

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				if err := a.service.EventRepository().DeleteOld(cleanupCtx); err != nil {
					log.Error().Err(err).Msg("failed to cleanup old events")
				}
			}
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
