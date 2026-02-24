//go:build integration

package testinfra

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	Client *redis.Client
	Addr   string
}

func NewRedis(t testing.TB) *RedisContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := tcredis.Run(ctx, "redis:8-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("testinfra: start redis container: %v", err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("testinfra: get redis endpoint: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("testinfra: ping redis: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("testinfra: terminate redis container: %v", err)
		}
	})

	return &RedisContainer{
		Client: client,
		Addr:   endpoint,
	}
}
