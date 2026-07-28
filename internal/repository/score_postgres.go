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

type ScoreRepository interface {
	UpsertBatch(context.Context, []domain.Evaluation) ([]domain.Evaluation, error)
	ListByJuryID(context.Context, uuid.UUID) ([]domain.Evaluation, error)
	ListByJuryAndTeamID(context.Context, uuid.UUID, uuid.UUID) ([]domain.Evaluation, error)
	TeamTotal(context.Context, uuid.UUID) (domain.TeamScoreTotal, error)
	ListTeamTotals(context.Context) ([]domain.TeamScoreTotal, error)
}

type ScorePostgres struct {
	pool *pgxpool.Pool
}

func NewScorePostgres(pool *pgxpool.Pool) (*ScorePostgres, error) {
	if pool == nil {
		return nil, errors.New("score repository pool cannot be nil")
	}
	return &ScorePostgres{pool: pool}, nil
}

func (r *ScorePostgres) UpsertBatch(
	ctx context.Context,
	evaluations []domain.Evaluation,
) ([]domain.Evaluation, error) {
	if _, err := domain.JuryEvaluationTotal(evaluations); err != nil {
		return nil, err
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin score transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var isClosed bool
	if err := tx.QueryRow(ctx,
		`SELECT is_closed FROM evaluation_state WHERE singleton_id = 1 FOR SHARE`,
	).Scan(&isClosed); err != nil {
		return nil, fmt.Errorf("lock evaluation state: %w", err)
	}
	if isClosed {
		return nil, domain.ErrEvaluationLocked
	}
	var submitted bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM submissions WHERE team_id = $1)`,
		evaluations[0].TeamID,
	).Scan(&submitted); err != nil {
		return nil, fmt.Errorf("check team submission: %w", err)
	}
	if !submitted {
		return nil, domain.ErrSubmissionNotFound
	}

	persisted := make([]domain.Evaluation, 0, domain.CriterionCount)
	for _, evaluation := range evaluations {
		saved, err := upsertEvaluation(ctx, tx, evaluation)
		if err != nil {
			return nil, err
		}
		persisted = append(persisted, saved)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit score transaction: %w", err)
	}
	return persisted, nil
}

func (r *ScorePostgres) ListByJuryID(
	ctx context.Context,
	juryID uuid.UUID,
) ([]domain.Evaluation, error) {
	const query = `
		SELECT ` + evaluationColumns + `
		FROM evaluations
		WHERE jury_id = $1
		ORDER BY team_id, criterion_id
	`
	return collectEvaluations(ctx, r.pool, query, juryID)
}

func (r *ScorePostgres) ListByJuryAndTeamID(
	ctx context.Context,
	juryID, teamID uuid.UUID,
) ([]domain.Evaluation, error) {
	const query = `
		SELECT ` + evaluationColumns + `
		FROM evaluations
		WHERE jury_id = $1 AND team_id = $2
		ORDER BY criterion_id
	`
	return collectEvaluations(ctx, r.pool, query, juryID, teamID)
}

func (r *ScorePostgres) TeamTotal(ctx context.Context, teamID uuid.UUID) (domain.TeamScoreTotal, error) {
	const query = `
		SELECT
			$1::uuid,
			COALESCE(SUM(score), 0),
			COUNT(DISTINCT jury_id)
		FROM evaluations
		WHERE team_id = $1
	`
	var total domain.TeamScoreTotal
	var score int
	err := r.pool.QueryRow(ctx, query, teamID).Scan(
		&total.TeamID, &score, &total.EvaluatedByCount,
	)
	if err != nil {
		return domain.TeamScoreTotal{}, fmt.Errorf("get team score total: %w", err)
	}
	total.Total = domain.Score(score)
	return total, nil
}

func (r *ScorePostgres) ListTeamTotals(ctx context.Context) ([]domain.TeamScoreTotal, error) {
	const query = `
		SELECT t.id, COALESCE(SUM(e.score), 0), COUNT(DISTINCT e.jury_id)
		FROM teams t
		LEFT JOIN evaluations e ON e.team_id = t.id
		GROUP BY t.id
		ORDER BY COALESCE(SUM(e.score), 0) DESC, t.id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list team totals: %w", err)
	}
	defer rows.Close()
	totals := make([]domain.TeamScoreTotal, 0)
	for rows.Next() {
		var total domain.TeamScoreTotal
		var score int
		if err := rows.Scan(&total.TeamID, &score, &total.EvaluatedByCount); err != nil {
			return nil, fmt.Errorf("scan team total: %w", err)
		}
		total.Total = domain.Score(score)
		totals = append(totals, total)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team totals: %w", err)
	}
	return totals, nil
}

func upsertEvaluation(
	ctx context.Context,
	tx pgx.Tx,
	evaluation domain.Evaluation,
) (domain.Evaluation, error) {
	const query = `
		INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (jury_id, team_id, criterion_id)
		DO UPDATE SET score = EXCLUDED.score, updated_at = clock_timestamp()
		RETURNING ` + evaluationColumns
	saved, err := scanEvaluation(tx.QueryRow(ctx, query,
		evaluation.JuryID, evaluation.TeamID, evaluation.CriterionID, evaluation.Score,
	))
	if err != nil {
		return domain.Evaluation{}, fmt.Errorf("upsert evaluation: %w", err)
	}
	return saved, nil
}

func collectEvaluations(
	ctx context.Context,
	queryer interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	},
	query string,
	arguments ...any,
) ([]domain.Evaluation, error) {
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
	err := row.Scan(
		&evaluation.ID,
		&evaluation.JuryID,
		&evaluation.TeamID,
		&criterionID,
		&score,
		&evaluation.UpdatedAt,
	)
	evaluation.CriterionID = domain.CriterionID(criterionID)
	evaluation.Score = domain.Score(score)
	return evaluation, err
}
