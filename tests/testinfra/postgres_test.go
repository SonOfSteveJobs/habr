//go:build integration

package testinfra_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/SonOfSteveJobs/habr/tests/testinfra"
)

func TestNewPostgres(t *testing.T) {
	migrationsDir := filepath.Join(testinfra.ProjectRoot(t), "migrations/auth")
	pg := testinfra.NewPostgres(t, migrationsDir)

	ctx := context.Background()

	var result int
	if err := pg.Pool.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %d", result)
	}

	var exists bool
	err := pg.Pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')",
	).Scan(&exists)
	if err != nil {
		t.Fatalf("check users table: %v", err)
	}
	if !exists {
		t.Fatal("users table not found after migrations")
	}

	err = pg.Pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'outbox')",
	).Scan(&exists)
	if err != nil {
		t.Fatalf("check outbox table: %v", err)
	}
	if !exists {
		t.Fatal("outbox table not found after migrations")
	}
}
