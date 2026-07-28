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

type QueryRepository interface {
	ListJuryTeams(context.Context, uuid.UUID) ([]domain.JuryTeam, error)
	AdminStats(context.Context) (domain.AdminStats, error)
	SetEvaluationClosed(context.Context, uuid.UUID, bool) (domain.EvaluationState, error)
	ExportSummary(context.Context) ([]domain.ExportSummaryRow, error)
	ExportDetails(context.Context) ([]domain.ExportDetailRow, error)
}

type QueryPostgres struct {
	pool *pgxpool.Pool
}

func NewQueryPostgres(pool *pgxpool.Pool) (*QueryPostgres, error) {
	if pool == nil {
		return nil, errors.New("query repository pool cannot be nil")
	}
	return &QueryPostgres{pool: pool}, nil
}

func (r *QueryPostgres) ListJuryTeams(ctx context.Context, juryID uuid.UUID) ([]domain.JuryTeam, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			s.solution_url,
			COUNT(DISTINCT e.criterion_id) = $2,
			COUNT(DISTINCT tm.user_id),
			s.updated_at
		FROM submissions s
		JOIN teams t ON t.id = s.team_id
		JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN evaluations e ON e.team_id = t.id AND e.jury_id = $1
		GROUP BY t.id, t.name, s.solution_url, s.updated_at
		ORDER BY t.name
	`
	rows, err := r.pool.Query(ctx, query, juryID, domain.CriterionCount)
	if err != nil {
		return nil, fmt.Errorf("list jury teams: %w", err)
	}
	defer rows.Close()
	teams := make([]domain.JuryTeam, 0)
	for rows.Next() {
		var team domain.JuryTeam
		if err := rows.Scan(
			&team.TeamID, &team.TeamName, &team.SolutionURL, &team.EvaluatedByMe,
			&team.MembersCount, &team.SubmissionAt,
		); err != nil {
			return nil, fmt.Errorf("scan jury team: %w", err)
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jury teams: %w", err)
	}
	return teams, nil
}

func (r *QueryPostgres) AdminStats(ctx context.Context) (domain.AdminStats, error) {
	const query = `
		SELECT
			(SELECT COUNT(*) FROM users WHERE role = 'USER'),
			(SELECT COUNT(*) FROM teams),
			(SELECT COUNT(*) FROM submissions),
			(SELECT COUNT(*) FROM users WHERE role = 'JURY'),
			(SELECT is_closed FROM evaluation_state WHERE singleton_id = 1)
	`
	var stats domain.AdminStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.TotalTeams,
		&stats.SubmittedSolutions,
		&stats.TotalJuries,
		&stats.EvaluationsClosed,
	)
	if err != nil {
		return domain.AdminStats{}, fmt.Errorf("load admin stats: %w", err)
	}
	return stats, nil
}

func (r *QueryPostgres) SetEvaluationClosed(
	ctx context.Context,
	adminID uuid.UUID,
	closed bool,
) (domain.EvaluationState, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return domain.EvaluationState{}, fmt.Errorf("begin evaluation state transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var current bool
	if err := tx.QueryRow(ctx,
		`SELECT is_closed FROM evaluation_state WHERE singleton_id = 1 FOR UPDATE`,
	).Scan(&current); err != nil {
		return domain.EvaluationState{}, fmt.Errorf("lock evaluation state: %w", err)
	}
	if current != closed {
		const update = `
			UPDATE evaluation_state
			SET is_closed = $1,
				closed_at = CASE WHEN $1 THEN clock_timestamp() ELSE NULL END,
				closed_by = CASE WHEN $1 THEN $2::uuid ELSE NULL END,
				updated_at = clock_timestamp()
			WHERE singleton_id = 1
		`
		if _, err := tx.Exec(ctx, update, closed, adminID); err != nil {
			return domain.EvaluationState{}, fmt.Errorf("update evaluation state: %w", err)
		}
		action := domain.EvaluationOpened
		if closed {
			action = domain.EvaluationClosed
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO evaluation_state_events (action, admin_id) VALUES ($1, $2)`,
			action, adminID,
		); err != nil {
			return domain.EvaluationState{}, fmt.Errorf("append evaluation state event: %w", err)
		}
	}
	var state domain.EvaluationState
	if err := tx.QueryRow(ctx, `
		SELECT is_closed, closed_at, closed_by, updated_at
		FROM evaluation_state WHERE singleton_id = 1
	`).Scan(&state.IsClosed, &state.ClosedAt, &state.ClosedBy, &state.UpdatedAt); err != nil {
		return domain.EvaluationState{}, fmt.Errorf("read evaluation state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.EvaluationState{}, fmt.Errorf("commit evaluation state: %w", err)
	}
	return state, nil
}

func (r *QueryPostgres) ExportSummary(ctx context.Context) ([]domain.ExportSummaryRow, error) {
	const query = `
		WITH member_agg AS (
			SELECT tm.team_id, COUNT(*) AS total_members,
				STRING_AGG(u.full_name, ', ' ORDER BY u.full_name) AS members
			FROM team_members tm
			JOIN users u ON u.id = tm.user_id
			GROUP BY tm.team_id
		),
		score_agg AS (
			SELECT team_id, SUM(score) AS total_score,
				COUNT(DISTINCT jury_id) AS evaluated_by_count
			FROM evaluations
			GROUP BY team_id
		)
		SELECT
			t.id, t.name, captain.full_name, COALESCE(captain.telegram, ''),
			COALESCE(s.solution_url, 'НЕ СДАНО'),
			COALESCE(m.total_members, 0), COALESCE(m.members, ''),
			COALESCE(sc.total_score, 0), COALESCE(sc.evaluated_by_count, 0)
		FROM teams t
		JOIN users captain ON captain.id = t.captain_id
		LEFT JOIN submissions s ON s.team_id = t.id
		LEFT JOIN member_agg m ON m.team_id = t.id
		LEFT JOIN score_agg sc ON sc.team_id = t.id
		ORDER BY total_score DESC, t.name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("export summary query: %w", err)
	}
	defer rows.Close()
	result := make([]domain.ExportSummaryRow, 0)
	for rows.Next() {
		var row domain.ExportSummaryRow
		if err := rows.Scan(
			&row.TeamID, &row.TeamName, &row.CaptainName, &row.CaptainTelegram,
			&row.SolutionURL, &row.TotalMembers, &row.Members,
			&row.TotalScore, &row.EvaluatedByCount,
		); err != nil {
			return nil, fmt.Errorf("scan export summary: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *QueryPostgres) ExportDetails(ctx context.Context) ([]domain.ExportDetailRow, error) {
	const query = `
		SELECT t.name, u.full_name, e.criterion_id, e.score
		FROM evaluations e
		JOIN teams t ON t.id = e.team_id
		JOIN users u ON u.id = e.jury_id
		ORDER BY t.name, u.full_name, e.criterion_id
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("export detail query: %w", err)
	}
	defer rows.Close()
	result := make([]domain.ExportDetailRow, 0)
	for rows.Next() {
		var row domain.ExportDetailRow
		if err := rows.Scan(&row.TeamName, &row.JuryName, &row.CriterionID, &row.Score); err != nil {
			return nil, fmt.Errorf("scan export detail: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
