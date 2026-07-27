// Package repository provides PostgreSQL-backed data access implementations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const userColumns = `id, full_name, university, email, telegram, password_hash, role, team_id, created_at`

// UserRepository describes persistence operations used by authentication and
// participant profile services.
type UserRepository interface {
	Create(context.Context, domain.User) (domain.User, error)
	GetByID(context.Context, uuid.UUID) (domain.User, error)
	GetByEmail(context.Context, string) (domain.User, error)
	Update(context.Context, domain.User) (domain.User, error)
	UpdatePasswordHash(context.Context, uuid.UUID, string) error
	ListByTeamID(context.Context, uuid.UUID) ([]domain.User, error)
	Delete(context.Context, uuid.UUID) error
}

// UserPostgres is a PostgreSQL implementation of UserRepository.
type UserPostgres struct {
	pool *pgxpool.Pool
}

// NewUserPostgres creates a user repository backed by pool.
func NewUserPostgres(pool *pgxpool.Pool) (*UserPostgres, error) {
	if pool == nil {
		return nil, errors.New("user repository pool cannot be nil")
	}
	return &UserPostgres{pool: pool}, nil
}

// Create persists a participant or administrator account.
func (r *UserPostgres) Create(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		INSERT INTO users (full_name, university, email, telegram, password_hash, role, team_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + userColumns

	created, err := scanUser(r.pool.QueryRow(ctx, query,
		user.FullName,
		user.University,
		user.Email,
		user.Telegram,
		user.PasswordHash,
		user.Role,
		user.TeamID,
	))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return created, nil
}

// GetByID returns a user by its primary key.
func (r *UserPostgres) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return user, nil
}

// GetByEmail returns the full account record needed for authentication.
func (r *UserPostgres) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE email = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return user, nil
}

// Update persists all mutable user fields and returns the updated record.
func (r *UserPostgres) Update(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		UPDATE users
		SET full_name = $2,
			university = $3,
			email = $4,
			telegram = $5,
			password_hash = $6,
			role = $7,
			team_id = $8
		WHERE id = $1
		RETURNING ` + userColumns

	updated, err := scanUser(r.pool.QueryRow(ctx, query,
		user.ID,
		user.FullName,
		user.University,
		user.Email,
		user.Telegram,
		user.PasswordHash,
		user.Role,
		user.TeamID,
	))
	if err != nil {
		return domain.User{}, mapUserError(err)
	}
	return updated, nil
}

// UpdatePasswordHash changes only the password hash for an existing account.
func (r *UserPostgres) UpdatePasswordHash(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const query = `UPDATE users SET password_hash = $2 WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update user password hash: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// ListByTeamID returns all team members in creation order.
func (r *UserPostgres) ListByTeamID(ctx context.Context, teamID uuid.UUID) ([]domain.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE team_id = $1 ORDER BY created_at, id`

	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list users by team ID: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}
	return users, nil
}

// Delete permanently removes a user account.
func (r *UserPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
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
		&user.TeamID,
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
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return domain.ErrEmailAlreadyExists
	}
	return fmt.Errorf("user repository query: %w", err)
}
