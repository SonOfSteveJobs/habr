//go:build integration

package testinfra_test

import (
	"context"
	"testing"

	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

func TestNewRedis(t *testing.T) {
	rc := testinfra.NewRedis(t)

	ctx := context.Background()

	if err := rc.Client.Set(ctx, "test-key", "test-value", 0).Err(); err != nil {
		t.Fatalf("redis SET: %v", err)
	}

	val, err := rc.Client.Get(ctx, "test-key").Result()
	if err != nil {
		t.Fatalf("redis GET: %v", err)
	}
	if val != "test-value" {
		t.Fatalf("expected test-value, got %s", val)
	}
}
