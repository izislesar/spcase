package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const submissionColumns = `id, team_id, solution_url, updated_at`

type SubmissionRepository interface {
	GetByTeamID(context.Context, uuid.UUID) (domain.Submission, error)
	Upsert(context.Context, uuid.UUID, string, time.Time) (domain.Submission, error)
}

type SubmissionPostgres struct {
	pool *pgxpool.Pool
}

func NewSubmissionPostgres(pool *pgxpool.Pool) (*SubmissionPostgres, error) {
	if pool == nil {
		return nil, errors.New("submission repository pool cannot be nil")
	}
	return &SubmissionPostgres{pool: pool}, nil
}

func (r *SubmissionPostgres) GetByTeamID(ctx context.Context, teamID uuid.UUID) (domain.Submission, error) {
	const query = `SELECT ` + submissionColumns + ` FROM submissions WHERE team_id = $1`
	submission, err := scanSubmission(r.pool.QueryRow(ctx, query, teamID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Submission{}, domain.ErrSubmissionNotFound
	}
	if err != nil {
		return domain.Submission{}, fmt.Errorf("get submission: %w", err)
	}
	return submission, nil
}

func (r *SubmissionPostgres) Upsert(
	ctx context.Context,
	captainID uuid.UUID,
	solutionURL string,
	deadline time.Time,
) (domain.Submission, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return domain.Submission{}, fmt.Errorf("begin submission transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var teamID uuid.UUID
	const lockTeamQuery = `
		SELECT t.id
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = $1
		FOR UPDATE OF t
	`
	if err := tx.QueryRow(ctx, lockTeamQuery, captainID).Scan(&teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Submission{}, domain.ErrNoTeam
		}
		return domain.Submission{}, fmt.Errorf("lock submission team: %w", err)
	}
	var persistedCaptainID uuid.UUID
	var memberCount int
	var deadlinePassed bool
	const validateQuery = `
		SELECT
			t.captain_id,
			(SELECT COUNT(*) FROM team_members WHERE team_id = t.id),
			clock_timestamp() >= $2
		FROM teams t
		WHERE t.id = $1
	`
	if err := tx.QueryRow(ctx, validateQuery, teamID, deadline.UTC()).Scan(
		&persistedCaptainID, &memberCount, &deadlinePassed,
	); err != nil {
		return domain.Submission{}, fmt.Errorf("validate submission: %w", err)
	}
	if persistedCaptainID != captainID {
		return domain.Submission{}, domain.ErrNotTeamCaptain
	}
	if !domain.SubmissionAllowed(memberCount) {
		return domain.Submission{}, domain.ErrMinimumTwoMembers
	}
	if deadlinePassed {
		return domain.Submission{}, domain.ErrDeadlinePassed
	}

	const query = `
		INSERT INTO submissions (team_id, solution_url)
		VALUES ($1, $2)
		ON CONFLICT (team_id)
		DO UPDATE SET solution_url = EXCLUDED.solution_url, updated_at = clock_timestamp()
		RETURNING ` + submissionColumns
	submission, err := scanSubmission(tx.QueryRow(ctx, query, teamID, solutionURL))
	if err != nil {
		return domain.Submission{}, fmt.Errorf("upsert submission: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Submission{}, fmt.Errorf("commit submission: %w", err)
	}
	return submission, nil
}

func scanSubmission(row pgx.Row) (domain.Submission, error) {
	var submission domain.Submission
	err := row.Scan(
		&submission.ID, &submission.TeamID, &submission.SolutionURL, &submission.UpdatedAt,
	)
	return submission, err
}
