package main

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/pkg/grpcvalidate"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
	authgrpc "github.com/SonOfSteveJobs/habr/services/auth/internal/handler/grpc"
	userrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/user"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/service"
)

func main() {
	if err := config.Load(); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err.Error()))
	}

	cfg := config.AppConfig()

	logger.Init(cfg.Logger().Level(), cfg.Logger().AsJson())
	log := logger.Logger()

	closer.Listen(syscall.SIGINT, syscall.SIGTERM)

	pool, err := pgxpool.New(context.Background(), cfg.DBURI())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create connection pool")
	}
	closer.AddNamed("postgres", func(_ context.Context) error {
		pool.Close()
		return nil
	})

	userRepo := userrepo.New(pool)
	authService := service.New(userRepo)
	handler := authgrpc.New(authService)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcvalidate.UnaryInterceptor()),
	)
	authv1.RegisterAuthServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)
	closer.AddNamed("gRPC server", func(_ context.Context) error {
		grpcServer.GracefulStop()
		return nil
	})

	listener, err := net.Listen("tcp", cfg.GRPCPort())
	if err != nil {
		log.Fatal().Err(err).Str("port", cfg.GRPCPort()).Msg("failed to listen")
	}

	log.Info().Str("port", cfg.GRPCPort()).Msg("starting gRPC server")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("gRPC server failed")
	}

	closer.Wait()
}
