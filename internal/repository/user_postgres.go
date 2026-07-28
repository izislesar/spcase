// Package repository provides PostgreSQL-backed data access implementations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const userColumns = `id, full_name, university, email, telegram, password_hash, role, auth_version, disabled_at, created_at`

type UserRepository interface {
	Create(context.Context, domain.User) (domain.User, error)
	GetByID(context.Context, uuid.UUID) (domain.User, error)
	GetByEmail(context.Context, string) (domain.User, error)
	GetAccountProjection(context.Context, uuid.UUID) (domain.AccountProjection, error)
	UpdatePasswordHash(context.Context, uuid.UUID, string) error
	SetDisabled(context.Context, uuid.UUID, bool) error
	IncrementAuthVersion(context.Context, uuid.UUID) error
}

type UserPostgres struct {
	pool *pgxpool.Pool
}

func NewUserPostgres(pool *pgxpool.Pool) (*UserPostgres, error) {
	if pool == nil {
		return nil, errors.New("user repository pool cannot be nil")
	}
	return &UserPostgres{pool: pool}, nil
}

func (r *UserPostgres) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		INSERT INTO users (full_name, university, email, telegram, password_hash, role)
		VALUES ($1, $2, LOWER($3), $4, $5, $6)
		RETURNING ` + userColumns
	created, err := scanUser(r.pool.QueryRow(ctx, query,
		user.FullName, user.University, user.Email, user.Telegram, user.PasswordHash, user.Role,
	))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return created, nil
}

func (r *UserPostgres) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return user, nil
}

func (r *UserPostgres) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE LOWER(email) = LOWER($1)`
	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return user, nil
}

func (r *UserPostgres) GetAccountProjection(ctx context.Context, id uuid.UUID) (domain.AccountProjection, error) {
	const query = `SELECT id, role, auth_version, disabled_at FROM users WHERE id = $1`
	var projection domain.AccountProjection
	var role string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&projection.ID, &role, &projection.AuthVersion, &projection.DisabledAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AccountProjection{}, domain.ErrUserNotFound
		}
		return domain.AccountProjection{}, fmt.Errorf("get account projection: %w", err)
	}
	projection.Role = domain.Role(role)
	return projection, nil
}

func (r *UserPostgres) UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const query = `
		UPDATE users
		SET password_hash = $2, auth_version = auth_version + 1
		WHERE id = $1
	`
	return requireUser(r.pool.Exec(ctx, query, id, passwordHash))
}

func (r *UserPostgres) SetDisabled(ctx context.Context, id uuid.UUID, disabled bool) error {
	const query = `
		UPDATE users
		SET disabled_at = CASE WHEN $2 THEN clock_timestamp() ELSE NULL END,
			auth_version = auth_version + 1
		WHERE id = $1
	`
	return requireUser(r.pool.Exec(ctx, query, id, disabled))
}

func (r *UserPostgres) IncrementAuthVersion(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE users SET auth_version = auth_version + 1 WHERE id = $1`
	return requireUser(r.pool.Exec(ctx, query, id))
}

func requireUser(commandTag pgconn.CommandTag, err error) error {
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	var role string
	if err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.University,
		&user.Email,
		&user.Telegram,
		&user.PasswordHash,
		&role,
		&user.AuthVersion,
		&user.DisabledAt,
		&user.CreatedAt,
	); err != nil {
		return domain.User{}, err
	}
	user.Role = domain.Role(role)
	return user, nil
}

func mapUserError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUserNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" &&
		strings.Contains(strings.ToLower(postgresError.ConstraintName), "email") {
		return domain.ErrEmailAlreadyExists
	}
	return fmt.Errorf("user repository query: %w", err)
}
