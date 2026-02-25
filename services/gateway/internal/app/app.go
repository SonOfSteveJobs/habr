package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/gateway/internal/config"
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

func (a *App) initRouter(_ context.Context) error {
	r := chi.NewRouter()

	a.router = r
	return nil
}

func (a *App) initHTTPServer(_ context.Context) error {
	cfg := config.AppConfig()

	a.httpServer = &http.Server{
		Addr:    cfg.HTTPPort(),
		Handler: a.router,
	}

	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	return nil
}
