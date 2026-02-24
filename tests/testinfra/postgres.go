//go:build integration

package testinfra

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	Pool *pgxpool.Pool
	URI  string
}

func NewPostgres(t testing.TB, migrationsDir string) *PostgresContainer {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := postgres.Run(ctx, "postgres:17",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("testinfra: start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testinfra: get postgres connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("testinfra: connect to postgres: %v", err)
	}

	applyMigrations(t, pool, migrationsDir)

	t.Cleanup(func() {
		pool.Close()
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("testinfra: terminate postgres container: %v", err)
		}
	})

	return &PostgresContainer{
		Pool: pool,
		URI:  connStr,
	}
}

func applyMigrations(t testing.TB, pool *pgxpool.Pool, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("testinfra: read migrations dir %s: %v", dir, err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e.Name())
		}
	}
	sort.Strings(sqlFiles)

	ctx := context.Background()
	for _, name := range sqlFiles {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("testinfra: read migration %s: %v", name, err)
		}

		upSQL := extractGooseUp(string(data))
		if upSQL == "" {
			continue
		}

		if _, err := pool.Exec(ctx, upSQL); err != nil {
			t.Fatalf("testinfra: apply migration %s: %v", name, err)
		}
	}
}

func extractGooseUp(content string) string {
	const upMarker = "-- +goose Up"
	const downMarker = "-- +goose Down"

	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return ""
	}

	sql := content[upIdx+len(upMarker):]

	if downIdx := strings.Index(sql, downMarker); downIdx != -1 {
		sql = sql[:downIdx]
	}

	return strings.TrimSpace(sql)
}
