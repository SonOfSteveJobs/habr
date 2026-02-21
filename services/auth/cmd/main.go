package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
	authgrpc "github.com/SonOfSteveJobs/habr/services/auth/internal/handler/grpc"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	handler := authgrpc.New()

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Str("port", cfg.GRPCPort).Msg("failed to listen")
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		logger.Info().Str("signal", sig.String()).Msg("shutting down")
		grpcServer.GracefulStop()
	}()

	logger.Info().Str("port", cfg.GRPCPort).Msg("starting gRPC server")

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server failed")
	}
}
