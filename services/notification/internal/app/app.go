package app

import (
	"context"
	"time"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/pkg/metrics"
	"github.com/SonOfSteveJobs/habr/pkg/tracing"
	"github.com/SonOfSteveJobs/habr/services/notification/internal/config"
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
		cfg := config.AppConfig()
		ticker := time.NewTicker(cfg.CleanupInterval())
		defer ticker.Stop()

		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := a.service.EventRepository().DeleteOld(ctx, cfg.RetentionPeriod()); err != nil {
					log.Error().Err(err).Msg("failed to cleanup old events")
				}
				cancel()
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
		{"tracing", a.initTracing},
		{"otel-logger", a.initOTelLogger},
		{"metrics", a.initMetrics},
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

func (a *App) initTracing(ctx context.Context) error {
	if err := tracing.InitTracer(ctx, config.AppConfig().Tracing()); err != nil {
		return err
	}

	closer.AddNamed("tracing", tracing.ShutdownTracer)

	return nil
}

func (a *App) initOTelLogger(ctx context.Context) error {
	if err := logger.EnableOTel(ctx, config.AppConfig().Tracing()); err != nil {
		return err
	}

	closer.AddNamed("otel-logger", logger.ShutdownOTelLogger)

	return nil
}

func (a *App) initMetrics(ctx context.Context) error {
	if err := metrics.InitMeter(ctx, config.AppConfig().Tracing()); err != nil {
		return err
	}

	closer.AddNamed("metrics", metrics.ShutdownMeter)

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
