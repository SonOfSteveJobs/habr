package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
	authgrpc "github.com/SonOfSteveJobs/habr/services/auth/internal/handler/grpc"
)

func main() {
	if err := config.Load(); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err.Error()))
	}

	cfg := config.AppConfig()

	logger.Init(cfg.Logger().Level(), cfg.Logger().AsJson())
	log := logger.Logger()

	handler := authgrpc.New()

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", cfg.GRPCPort())
	if err != nil {
		log.Fatal().Err(err).Str("port", cfg.GRPCPort()).Msg("failed to listen")
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		log.Info().Str("signal", sig.String()).Msg("shutting down")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("port", cfg.GRPCPort()).Msg("starting gRPC server")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("gRPC server failed")
	}
}
