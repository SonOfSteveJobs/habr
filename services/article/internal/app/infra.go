package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/SonOfSteveJobs/habr/pkg/closer"
	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/article/internal/config"
)

type infraContainer struct {
	pgPool      *pgxpool.Pool
	redisClient *redis.Client
	txManager   *transaction.Manager
}

func newInfraContainer(ctx context.Context) (*infraContainer, error) {
	c := &infraContainer{}

	if err := c.initPgPool(ctx); err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	if err := c.initRedisClient(ctx); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}

	c.txManager = transaction.New(c.pgPool)

	return c, nil
}

func (c *infraContainer) PgPool() *pgxpool.Pool           { return c.pgPool }
func (c *infraContainer) RedisClient() *redis.Client      { return c.redisClient }
func (c *infraContainer) TxManager() *transaction.Manager { return c.txManager }

func (c *infraContainer) initPgPool(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, config.AppConfig().DBURI())
	if err != nil {
		return err
	}
	closer.AddNamed("postgres", func(_ context.Context) error {
		pool.Close()
		return nil
	})

	c.pgPool = pool
	return nil
}

func (c *infraContainer) initRedisClient(ctx context.Context) error {
	client := redis.NewClient(&redis.Options{
		Addr: config.AppConfig().RedisAddr(),
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	closer.AddNamed("redis", func(_ context.Context) error {
		return client.Close()
	})

	c.redisClient = client
	return nil
}
