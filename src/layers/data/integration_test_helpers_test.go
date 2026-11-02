//go:build integration

package data

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupIntegrationDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Some local/rootless Docker setups cannot run Ryuk; allow direct container lifecycle management.
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	t.Setenv("TESTCONTAINERS_CHECKS_DISABLE", "true")

	ctx := context.Background()
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("nam_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres testcontainer: %v", err)
	}

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to build postgres connection string: %v", err)
	}

	if err := AutoMigrate(dsn); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to open pgx pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping postgres: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}
