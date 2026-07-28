package postgres

import (
	"testing"
	"time"

	"spcase.ru/backend/internal/config"
)

func TestPoolConfigurationAppliesPostgresTimeouts(t *testing.T) {
	t.Parallel()

	database := config.DatabaseConfig{
		Host:             "127.0.0.1",
		Port:             5432,
		User:             "postgres",
		Password:         "secret",
		Name:             "spcase",
		StatementTimeout: 15 * time.Second,
		LockTimeout:      5 * time.Second,
	}

	poolConfig, err := poolConfiguration(database)
	if err != nil {
		t.Fatalf("poolConfiguration() error = %v", err)
	}
	if got, want := poolConfig.ConnConfig.RuntimeParams["statement_timeout"], "15000"; got != want {
		t.Fatalf("statement_timeout = %q, want %q", got, want)
	}
	if got, want := poolConfig.ConnConfig.RuntimeParams["lock_timeout"], "5000"; got != want {
		t.Fatalf("lock_timeout = %q, want %q", got, want)
	}
}
