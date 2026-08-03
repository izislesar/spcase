//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/config"
)

const integrationDatabaseEnvironment = "SPCASE_TEST_APP_DATABASE_URL"

func TestStatementTimeoutCancelsLongRunningQuery(t *testing.T) {
	pool := timeoutTestPool(t, 50*time.Millisecond, time.Second)

	_, err := pool.Exec(context.Background(), `SELECT pg_sleep(0.25)`)
	requirePostgresErrorCode(t, err, "57014")
}

func TestLockTimeoutCancelsLockWait(t *testing.T) {
	pool := timeoutTestPool(t, time.Second, 50*time.Millisecond)
	ctx := context.Background()

	blocker, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire blocking connection: %v", err)
	}
	defer blocker.Release()
	waiter, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire waiting connection: %v", err)
	}
	defer waiter.Release()

	lockID := time.Now().UnixNano()
	if _, err := blocker.Exec(ctx, `SELECT pg_advisory_lock($1)`, lockID); err != nil {
		t.Fatalf("acquire advisory lock: %v", err)
	}
	defer blocker.Exec(ctx, `SELECT pg_advisory_unlock($1)`, lockID)

	_, err = waiter.Exec(ctx, `SELECT pg_advisory_lock($1)`, lockID)
	requirePostgresErrorCode(t, err, "55P03")
}

func timeoutTestPool(
	t *testing.T,
	statementTimeout time.Duration,
	lockTimeout time.Duration,
) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv(integrationDatabaseEnvironment)
	if databaseURL == "" {
		t.Skipf("set %s to run PostgreSQL timeout integration tests", integrationDatabaseEnvironment)
	}
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse integration database URL: %v", err)
	}
	applySessionTimeouts(poolConfig, config.DatabaseConfig{
		StatementTimeout: statementTimeout,
		LockTimeout:      lockTimeout,
	})
	poolConfig.MaxConns = 2
	poolConfig.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("create timeout test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping integration database: %v", err)
	}
	var role string
	if err := pool.QueryRow(ctx, `SELECT current_user`).Scan(&role); err != nil {
		t.Fatalf("verify timeout test database role: %v", err)
	}
	if role != "spcase_app" {
		t.Fatalf("timeout test pool connected as %q, want spcase_app", role)
	}
	return pool
}

func requirePostgresErrorCode(t *testing.T, err error, expected string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("error = %v, want PostgreSQL error %s", err, expected)
	}
	if postgresError.Code != expected {
		t.Fatalf("PostgreSQL error code = %q, want %q: %v", postgresError.Code, expected, err)
	}
}
