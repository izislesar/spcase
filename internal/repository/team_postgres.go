package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const teamColumns = `id, name, invite_code, captain_id, created_at`

// TeamRepository describes persistence operations for teams and their members.
type TeamRepository interface {
	Create(context.Context, domain.Team) (domain.Team, error)
	GetByID(context.Context, uuid.UUID) (domain.Team, error)
	GetByInviteCode(context.Context, string) (domain.Team, error)
	ListMembers(context.Context, uuid.UUID) ([]domain.TeamMember, error)
	Join(context.Context, uuid.UUID, string) (domain.Team, error)
	Leave(context.Context, uuid.UUID) error
	Kick(context.Context, uuid.UUID, uuid.UUID) error
	TransferOwnership(context.Context, uuid.UUID, uuid.UUID) error
	Disband(context.Context, uuid.UUID) error
}

// TeamPostgres is a PostgreSQL implementation of TeamRepository.
type TeamPostgres struct {
	pool *pgxpool.Pool
}

// NewTeamPostgres creates a team repository backed by pool.
func NewTeamPostgres(pool *pgxpool.Pool) (*TeamPostgres, error) {
	if pool == nil {
		return nil, errors.New("team repository pool cannot be nil")
	}
	return &TeamPostgres{pool: pool}, nil
}

// Create atomically creates a team and assigns its captain to it.
func (r *TeamPostgres) Create(ctx context.Context, team domain.Team) (domain.Team, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback(ctx)

	teamID, err := lockUserTeamID(ctx, tx, team.CaptainID)
	if err != nil {
		return domain.Team{}, err
	}
	if teamID != nil {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}

	const query = `
		INSERT INTO teams (name, invite_code, captain_id)
		VALUES ($1, $2, $3)
		RETURNING ` + teamColumns
	created, err := scanTeam(tx.QueryRow(ctx, query, team.Name, team.InviteCode, team.CaptainID))
	if err != nil {
		return domain.Team{}, fmt.Errorf("create team: %w", err)
	}

	const assignCaptainQuery = `UPDATE users SET team_id = $2 WHERE id = $1 AND team_id IS NULL`
	commandTag, err := tx.Exec(ctx, assignCaptainQuery, team.CaptainID, created.ID)
	if err != nil {
		return domain.Team{}, fmt.Errorf("assign team captain: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, fmt.Errorf("commit team creation: %w", err)
	}
	return created, nil
}

// GetByID returns a team by its primary key.
func (r *TeamPostgres) GetByID(ctx context.Context, id uuid.UUID) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE id = $1`
	team, err := scanTeam(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("get team by ID: %w", err)
	}
	return team, nil
}

// GetByInviteCode returns a team by its invite code.
func (r *TeamPostgres) GetByInviteCode(ctx context.Context, inviteCode string) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE invite_code = $1`
	team, err := scanTeam(r.pool.QueryRow(ctx, query, inviteCode))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrInvalidInviteCode
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("get team by invite code: %w", err)
	}
	return team, nil
}

// ListMembers returns a team's roster with captain markers.
func (r *TeamPostgres) ListMembers(ctx context.Context, teamID uuid.UUID) ([]domain.TeamMember, error) {
	const query = `
		SELECT id, full_name, telegram, id = $2 AS is_captain
		FROM users
		WHERE team_id = $1
		ORDER BY created_at, id
	`

	rows, err := r.pool.Query(ctx, query, teamID, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.ID, &member.FullName, &member.Telegram, &member.IsCaptain); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}
	return members, nil
}

// Join atomically assigns a user without a team to the team identified by code.
func (r *TeamPostgres) Join(ctx context.Context, userID uuid.UUID, inviteCode string) (domain.Team, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback(ctx)

	userTeamID, err := lockUserTeamID(ctx, tx, userID)
	if err != nil {
		return domain.Team{}, err
	}
	if userTeamID != nil {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}
	team, err := lockTeamByInviteCode(ctx, tx, inviteCode)
	if err != nil {
		return domain.Team{}, err
	}

	memberCount, err := countTeamMembers(ctx, tx, team.ID)
	if err != nil {
		return domain.Team{}, err
	}
	if !domain.HasCapacity(memberCount) {
		return domain.Team{}, domain.ErrTeamFull
	}

	const query = `UPDATE users SET team_id = $2 WHERE id = $1 AND team_id IS NULL`
	commandTag, err := tx.Exec(ctx, query, userID, team.ID)
	if err != nil {
		return domain.Team{}, fmt.Errorf("join team: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, fmt.Errorf("commit team join: %w", err)
	}
	return team, nil
}

// Leave atomically removes a non-captain member from their current team.
func (r *TeamPostgres) Leave(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	teamID, err := lockUserTeamID(ctx, tx, userID)
	if err != nil {
		return err
	}
	if teamID == nil {
		return domain.ErrNoTeam
	}
	team, err := lockTeamByID(ctx, tx, *teamID)
	if err != nil {
		return err
	}
	if team.IsCaptain(userID) {
		return domain.ErrCaptainCannotLeave
	}

	const query = `UPDATE users SET team_id = NULL WHERE id = $1 AND team_id = $2`
	commandTag, err := tx.Exec(ctx, query, userID, *teamID)
	if err != nil {
		return fmt.Errorf("leave team: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.ErrNoTeam
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit team leave: %w", err)
	}
	return nil
}

// Kick atomically removes a member when requested by that team's captain.
func (r *TeamPostgres) Kick(ctx context.Context, captainID, memberID uuid.UUID) error {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	teamID, err := lockUserTeamID(ctx, tx, captainID)
	if err != nil {
		return err
	}
	if teamID == nil {
		return domain.ErrNoTeam
	}
	team, err := lockTeamByID(ctx, tx, *teamID)
	if err != nil {
		return err
	}
	if !team.IsCaptain(captainID) {
		return domain.ErrNotTeamCaptain
	}
	if captainID == memberID {
		return domain.ErrCaptainCannotBeKicked
	}

	memberTeamID, err := lockUserTeamID(ctx, tx, memberID)
	if err != nil {
		return err
	}
	if memberTeamID == nil || *memberTeamID != team.ID {
		return domain.ErrTeamMemberNotFound
	}

	const query = `UPDATE users SET team_id = NULL WHERE id = $1 AND team_id = $2`
	commandTag, err := tx.Exec(ctx, query, memberID, team.ID)
	if err != nil {
		return fmt.Errorf("kick team member: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.ErrTeamMemberNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit team kick: %w", err)
	}
	return nil
}

// TransferOwnership atomically assigns captaincy to another current member.
func (r *TeamPostgres) TransferOwnership(ctx context.Context, captainID, newCaptainID uuid.UUID) error {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	team, err := lockCaptainTeam(ctx, tx, captainID)
	if err != nil {
		return err
	}
	if newCaptainID == captainID {
		return domain.ErrCaptainCannotBeKicked
	}
	newCaptainTeamID, err := lockUserTeamID(ctx, tx, newCaptainID)
	if err != nil {
		return err
	}
	if newCaptainTeamID == nil || *newCaptainTeamID != team.ID {
		return domain.ErrTeamMemberNotFound
	}

	const query = `UPDATE teams SET captain_id = $2 WHERE id = $1`
	if _, err := tx.Exec(ctx, query, team.ID, newCaptainID); err != nil {
		return fmt.Errorf("transfer team ownership: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ownership transfer: %w", err)
	}
	return nil
}

// Disband atomically deletes a captain's team. The users.team_id foreign key
// releases all members through its ON DELETE SET NULL constraint.
func (r *TeamPostgres) Disband(ctx context.Context, captainID uuid.UUID) error {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	team, err := lockCaptainTeam(ctx, tx, captainID)
	if err != nil {
		return err
	}
	const query = `DELETE FROM teams WHERE id = $1 AND captain_id = $2`
	commandTag, err := tx.Exec(ctx, query, team.ID, captainID)
	if err != nil {
		return fmt.Errorf("disband team: %w", err)
	}
	if commandTag.RowsAffected() != 1 {
		return domain.ErrNotTeamCaptain
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit team disband: %w", err)
	}
	return nil
}

func (r *TeamPostgres) beginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin team transaction: %w", err)
	}
	return tx, nil
}

func lockCaptainTeam(ctx context.Context, tx pgx.Tx, captainID uuid.UUID) (domain.Team, error) {
	teamID, err := lockUserTeamID(ctx, tx, captainID)
	if err != nil {
		return domain.Team{}, err
	}
	if teamID == nil {
		return domain.Team{}, domain.ErrNoTeam
	}
	team, err := lockTeamByID(ctx, tx, *teamID)
	if err != nil {
		return domain.Team{}, err
	}
	if !team.IsCaptain(captainID) {
		return domain.Team{}, domain.ErrNotTeamCaptain
	}
	return team, nil
}

func lockTeamByInviteCode(ctx context.Context, tx pgx.Tx, inviteCode string) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE invite_code = $1 FOR UPDATE`
	team, err := scanTeam(tx.QueryRow(ctx, query, inviteCode))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrInvalidInviteCode
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("lock team by invite code: %w", err)
	}
	return team, nil
}

func lockTeamByID(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE id = $1 FOR UPDATE`
	team, err := scanTeam(tx.QueryRow(ctx, query, teamID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("lock team by ID: %w", err)
	}
	return team, nil
}

func lockUserTeamID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*uuid.UUID, error) {
	const query = `SELECT team_id FROM users WHERE id = $1 FOR UPDATE`
	var teamID *uuid.UUID
	if err := tx.QueryRow(ctx, query, userID).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("lock user: %w", err)
	}
	return teamID, nil
}

func countTeamMembers(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM users WHERE team_id = $1`
	var count int
	if err := tx.QueryRow(ctx, query, teamID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count team members: %w", err)
	}
	return count, nil
}

func scanTeam(row pgx.Row) (domain.Team, error) {
	var team domain.Team
	if err := row.Scan(&team.ID, &team.Name, &team.InviteCode, &team.CaptainID, &team.CreatedAt); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}
