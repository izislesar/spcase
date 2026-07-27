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

const evaluationColumns = `id, jury_id, team_id, criterion_id, score, updated_at`

// TeamScoreTotal is the aggregate score of one team across every jury member
// and evaluation criterion.
type TeamScoreTotal struct {
	TeamID uuid.UUID
	Total  domain.Score
}

// ScoreRepository describes persistence operations for jury evaluations.
type ScoreRepository interface {
	Upsert(context.Context, domain.Evaluation) (domain.Evaluation, error)
	UpsertBatch(context.Context, []domain.Evaluation) ([]domain.Evaluation, error)
	ListByJuryID(context.Context, uuid.UUID) ([]domain.Evaluation, error)
	ListByJuryAndTeamID(context.Context, uuid.UUID, uuid.UUID) ([]domain.Evaluation, error)
	TeamTotal(context.Context, uuid.UUID) (domain.Score, error)
	ListTeamTotals(context.Context) ([]TeamScoreTotal, error)
}

// ScorePostgres is a PostgreSQL implementation of ScoreRepository.
type ScorePostgres struct {
	pool *pgxpool.Pool
}

// NewScorePostgres creates a score repository backed by pool.
func NewScorePostgres(pool *pgxpool.Pool) (*ScorePostgres, error) {
	if pool == nil {
		return nil, errors.New("score repository pool cannot be nil")
	}
	return &ScorePostgres{pool: pool}, nil
}

// Upsert saves one criterion score, replacing the previous score from the
// same jury member for the same team and criterion when it exists.
func (r *ScorePostgres) Upsert(ctx context.Context, evaluation domain.Evaluation) (domain.Evaluation, error) {
	evaluations, err := r.UpsertBatch(ctx, []domain.Evaluation{evaluation})
	if err != nil {
		return domain.Evaluation{}, err
	}
	return evaluations[0], nil
}

// UpsertBatch atomically saves a set of scores. A duplicate criterion in the
// supplied set is rejected instead of making its final value order-dependent.
func (r *ScorePostgres) UpsertBatch(ctx context.Context, evaluations []domain.Evaluation) ([]domain.Evaluation, error) {
	if len(evaluations) == 0 {
		return []domain.Evaluation{}, nil
	}
	if err := validateEvaluations(evaluations); err != nil {
		return nil, err
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin score upsert transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	persisted := make([]domain.Evaluation, 0, len(evaluations))
	for _, evaluation := range evaluations {
		saved, err := upsertEvaluation(ctx, tx, evaluation)
		if err != nil {
			return nil, err
		}
		persisted = append(persisted, saved)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit score upsert transaction: %w", err)
	}
	return persisted, nil
}

// ListByJuryID returns only evaluations authored by the requested jury member.
func (r *ScorePostgres) ListByJuryID(ctx context.Context, juryID uuid.UUID) ([]domain.Evaluation, error) {
	const query = `
		SELECT ` + evaluationColumns + `
		FROM evaluations
		WHERE jury_id = $1
		ORDER BY team_id, criterion_id
	`
	return collectEvaluations(ctx, r.pool, query, juryID)
}

// ListByJuryAndTeamID returns one jury member's evaluations for a single team.
func (r *ScorePostgres) ListByJuryAndTeamID(ctx context.Context, juryID, teamID uuid.UUID) ([]domain.Evaluation, error) {
	const query = `
		SELECT ` + evaluationColumns + `
		FROM evaluations
		WHERE jury_id = $1 AND team_id = $2
		ORDER BY criterion_id
	`
	return collectEvaluations(ctx, r.pool, query, juryID, teamID)
}

// TeamTotal returns the sum of all saved scores for one team. A team with no
// evaluations has a total of zero.
func (r *ScorePostgres) TeamTotal(ctx context.Context, teamID uuid.UUID) (domain.Score, error) {
	const query = `SELECT COALESCE(SUM(score), 0) FROM evaluations WHERE team_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, query, teamID).Scan(&total); err != nil {
		return 0, fmt.Errorf("get team score total: %w", err)
	}
	return domain.Score(total), nil
}

// ListTeamTotals returns score aggregates for every team with at least one
// saved evaluation, ranked from the highest score to the lowest.
func (r *ScorePostgres) ListTeamTotals(ctx context.Context) ([]TeamScoreTotal, error) {
	const query = `
		SELECT team_id, SUM(score) AS total
		FROM evaluations
		GROUP BY team_id
		ORDER BY total DESC, team_id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list team score totals: %w", err)
	}
	defer rows.Close()

	totals := make([]TeamScoreTotal, 0)
	for rows.Next() {
		var total TeamScoreTotal
		var score int
		if err := rows.Scan(&total.TeamID, &score); err != nil {
			return nil, fmt.Errorf("scan team score total: %w", err)
		}
		total.Total = domain.Score(score)
		totals = append(totals, total)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team score totals: %w", err)
	}
	return totals, nil
}

func upsertEvaluation(ctx context.Context, tx pgx.Tx, evaluation domain.Evaluation) (domain.Evaluation, error) {
	const query = `
		INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (jury_id, team_id, criterion_id)
		DO UPDATE SET score = EXCLUDED.score, updated_at = CURRENT_TIMESTAMP
		RETURNING ` + evaluationColumns

	saved, err := scanEvaluation(tx.QueryRow(ctx, query,
		evaluation.JuryID,
		evaluation.TeamID,
		evaluation.CriterionID,
		evaluation.Score,
	))
	if err != nil {
		return domain.Evaluation{}, fmt.Errorf("upsert evaluation: %w", err)
	}
	return saved, nil
}

func collectEvaluations(ctx context.Context, queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, query string, arguments ...any) ([]domain.Evaluation, error) {
	rows, err := queryer.Query(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query evaluations: %w", err)
	}
	defer rows.Close()

	evaluations := make([]domain.Evaluation, 0)
	for rows.Next() {
		evaluation, err := scanEvaluation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan evaluation: %w", err)
		}
		evaluations = append(evaluations, evaluation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evaluations: %w", err)
	}
	return evaluations, nil
}

func scanEvaluation(row pgx.Row) (domain.Evaluation, error) {
	var evaluation domain.Evaluation
	var criterionID, score int
	if err := row.Scan(
		&evaluation.ID,
		&evaluation.JuryID,
		&evaluation.TeamID,
		&criterionID,
		&score,
		&evaluation.UpdatedAt,
	); err != nil {
		return domain.Evaluation{}, err
	}
	evaluation.CriterionID = domain.CriterionID(criterionID)
	evaluation.Score = domain.Score(score)
	return evaluation, nil
}

func validateEvaluations(evaluations []domain.Evaluation) error {
	criteria := make(map[domain.CriterionID]struct{}, len(evaluations))
	for _, evaluation := range evaluations {
		if evaluation.JuryID == uuid.Nil || evaluation.TeamID == uuid.Nil || !evaluation.IsValid() {
			return domain.ErrInvalidEvaluation
		}
		if _, exists := criteria[evaluation.CriterionID]; exists {
			return domain.ErrInvalidEvaluation
		}
		criteria[evaluation.CriterionID] = struct{}{}
	}
	return nil
}
