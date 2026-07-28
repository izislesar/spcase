// Package postgres provides PostgreSQL connection-pool initialization.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/config"
)

const (
	connectTimeout = 5 * time.Second
	pingTimeout    = 5 * time.Second
)

// New creates a PostgreSQL pool, validates connectivity with Ping, and closes
// the pool if initialization fails. The caller owns the returned pool and must
// close it during graceful shutdown.
func New(ctx context.Context, db config.DatabaseConfig) (*pgxpool.Pool, error) {
	if ctx == nil {
		return nil, errors.New("postgres initialization context cannot be nil")
	}

	poolConfig, err := poolConfiguration(db)
	if err != nil {
		return nil, err
	}

	connectCtx, cancelConnect := context.WithTimeout(ctx, connectTimeout)
	defer cancelConnect()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, pingTimeout)
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}

func poolConfiguration(db config.DatabaseConfig) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(connectionString(db))
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}
	poolConfig.ConnConfig.ConnectTimeout = connectTimeout
	applySessionTimeouts(poolConfig, db)
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute
	return poolConfig, nil
}

func applySessionTimeouts(poolConfig *pgxpool.Config, db config.DatabaseConfig) {
	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = durationMilliseconds(db.StatementTimeout)
	poolConfig.ConnConfig.RuntimeParams["lock_timeout"] = durationMilliseconds(db.LockTimeout)
}

func durationMilliseconds(duration time.Duration) string {
	return strconv.FormatInt(duration.Milliseconds(), 10)
}

func connectionString(db config.DatabaseConfig) string {
	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(db.User, db.Password),
		Host:   net.JoinHostPort(db.Host, strconv.Itoa(db.Port)),
		Path:   db.Name,
	}
	return connectionURL.String()
}
