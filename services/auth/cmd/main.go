package main

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	authv1 "github.com/SonOfSteveJobs/habr/pkg/gen/auth/v1"
	"github.com/SonOfSteveJobs/habr/pkg/grpcvalidate"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/producer"
	"github.com/SonOfSteveJobs/habr/pkg/logger"
	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/config"
	authgrpc "github.com/SonOfSteveJobs/habr/services/auth/internal/handler/grpc"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/outbox"
	outboxrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/outbox"
	tokenrepo "github.com/SonOfSteveJobs/habr/services/auth/internal/repository/token"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr(),
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	closer.AddNamed("redis", func(_ context.Context) error {
		return redisClient.Close()
	})

	txManager := transaction.New(pool)
	outboxRepo := outboxrepo.New(txManager)

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

		if err := outboxRepo.MarkSent(context.Background(), uid); err != nil {
			log.Error().Err(err).Str("event_id", eventID).Msg("outbox: mark sent failed")
		}
	}

	kafkaCfg := producer.NewAsyncConfig(producer.WithIdempotent())
	saramaProducer, err := sarama.NewAsyncProducer(cfg.Kafka().Brokers(), kafkaCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka producer")
	}
	kafkaProducer := producer.NewAsync(saramaProducer, cfg.Kafka().Topic(), onSuccess)
	closer.AddNamed("kafka producer", func(_ context.Context) error {
		return kafkaProducer.Close()
	})

	relay := outbox.NewRelay(outboxRepo, kafkaProducer)
	relayCtx, relayCancel := context.WithCancel(context.Background())
	go relay.Run(relayCtx)
	closer.AddNamed("outbox relay", func(_ context.Context) error {
		relayCancel()
		return nil
	})

	userRepo := userrepo.New(txManager)
	tokenRepo := tokenrepo.New(redisClient)
	authService := service.New(
		userRepo, tokenRepo, outboxRepo, txManager,
		cfg.JWTSecret(), cfg.Kafka().Topic(),
		cfg.AccessTokenTTL(), cfg.RefreshTokenTTL(),
	)
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

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Error().Err(err).Msg("gRPC server failed")
		}
	}()
	closer.Wait()
}
