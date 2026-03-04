package app

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	articlev1 "github.com/SonOfSteveJobs/habr/pkg/gen/article/v1"
	"github.com/SonOfSteveJobs/habr/pkg/grpcvalidate"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/article/internal/config"
)

type App struct {
	infra   *infraContainer
	service *serviceContainer

	grpcServer *grpc.Server
	listener   net.Listener
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

	log.Info().Str("port", cfg.GRPCPort()).Msg("starting gRPC server")

	go func() {
		if err := a.grpcServer.Serve(a.listener); err != nil {
			log.Error().Err(err).Msg("gRPC server failed")
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
		{"infra: postgres, redis", a.initInfra},
		{"service: repositories, services", a.initService},
		{"gRPC server", a.initGRPCServer},
		{"listener", a.initListener},
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

func (a *App) initGRPCServer(_ context.Context) error {
	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(grpcvalidate.UnaryInterceptor()),
	)
	articlev1.RegisterArticleServiceServer(a.grpcServer, a.service.Handler())
	reflection.Register(a.grpcServer)
	closer.AddNamed("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	return nil
}

func (a *App) initListener(_ context.Context) error {
	listener, err := net.Listen("tcp", config.AppConfig().GRPCPort())
	if err != nil {
		return err
	}

	a.listener = listener

	return nil
}
