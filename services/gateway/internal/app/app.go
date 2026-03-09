package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	gatewayv1 "github.com/SonOfSteveJobs/habr/pkg/gen/gateway/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/pkg/metrics"
	"github.com/SonOfSteveJobs/habr/pkg/tracing"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/config"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/handler/middleware"
)

type App struct {
	infra   *infraContainer
	service *serviceContainer

	router     chi.Router
	httpServer *http.Server
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
	cfg := config.AppConfig()

	log.Info().Str("port", cfg.HTTPPort()).Msg("starting HTTP server")

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP server failed")
		}
	}()

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
		{"infra: grpc connections", a.initInfra},
		{"service: clients", a.initService},
		{"HTTP router", a.initRouter},
		{"HTTP server", a.initHTTPServer},
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

func (a *App) initInfra(_ context.Context) error {
	infra, err := newInfraContainer()
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

func (a *App) initRouter(_ context.Context) error {
	cfg := config.AppConfig()
	r := chi.NewRouter()

	r.Use(tracing.HTTPMiddleware())
	r.Use(metrics.HTTPMiddleware())

	gatewayv1.HandlerWithOptions(a.service.Handler(), gatewayv1.ChiServerOptions{
		BaseRouter: r,
		Middlewares: []gatewayv1.MiddlewareFunc{
			middleware.Auth(cfg.JWTSecret()),
		},
	})

	a.router = r
	return nil
}

func (a *App) initHTTPServer(_ context.Context) error {
	cfg := config.AppConfig()

	a.httpServer = &http.Server{
		Addr:              cfg.HTTPPort(),
		Handler:           a.router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	return nil
}
