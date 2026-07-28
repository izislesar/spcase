//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const integrationDatabaseEnvironment = "SPCASE_TEST_DATABASE_URL"

var (
	integrationPool    *pgxpool.Pool
	integrationEnabled bool
	integrationSchema  string
)

func TestMain(m *testing.M) {
	databaseURL := os.Getenv(integrationDatabaseEnvironment)
	if databaseURL == "" {
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	adminConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		fatalIntegrationSetup("parse test database URL", err)
	}
	adminPool, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		fatalIntegrationSetup("connect to test database", err)
	}

	integrationSchema = "spcase_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quotedSchema := pgx.Identifier{integrationSchema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		adminPool.Close()
		fatalIntegrationSetup("create isolated test schema", err)
	}

	testConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		dropIntegrationSchema(ctx, adminPool, quotedSchema)
		fatalIntegrationSetup("parse isolated test database URL", err)
	}
	testConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	testConfig.ConnConfig.RuntimeParams["search_path"] = integrationSchema + ",public"
	testConfig.MaxConns = 12
	integrationPool, err = pgxpool.NewWithConfig(ctx, testConfig)
	if err != nil {
		dropIntegrationSchema(ctx, adminPool, quotedSchema)
		fatalIntegrationSetup("connect to isolated test schema", err)
	}
	if err := applyUpMigration(ctx, integrationPool, "00001_init_schema.sql"); err != nil {
		integrationPool.Close()
		dropIntegrationSchema(ctx, adminPool, quotedSchema)
		fatalIntegrationSetup("apply schema migration", err)
	}
	if err := applyUpMigration(ctx, integrationPool, "00002_add_indexes.sql"); err != nil {
		integrationPool.Close()
		dropIntegrationSchema(ctx, adminPool, quotedSchema)
		fatalIntegrationSetup("apply index migration", err)
	}
	integrationEnabled = true

	exitCode := m.Run()
	integrationPool.Close()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	dropIntegrationSchema(cleanupCtx, adminPool, quotedSchema)
	cleanupCancel()
	adminPool.Close()
	os.Exit(exitCode)
}

func fatalIntegrationSetup(operation string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "database integration setup: %s: %v\n", operation, err)
	os.Exit(1)
}

func dropIntegrationSchema(ctx context.Context, pool *pgxpool.Pool, quotedSchema string) {
	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "database integration cleanup: %v\n", err)
	}
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if !integrationEnabled {
		t.Skipf("set %s to run PostgreSQL integration tests", integrationDatabaseEnvironment)
	}
}

func applyUpMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	up, _, err := readMigrationSections(name)
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, up); err != nil {
		return fmt.Errorf("execute %s Up section: %w", name, err)
	}
	return nil
}

func applyDownMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	_, down, err := readMigrationSections(name)
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, down); err != nil {
		return fmt.Errorf("execute %s Down section: %w", name, err)
	}
	return nil
}

func readMigrationSections(name string) (string, string, error) {
	path := filepath.Join("..", "..", "migrations", name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read %s: %w", path, err)
	}
	parts := strings.SplitN(string(content), "-- +goose Down", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%s has no Goose Down section", path)
	}
	return parts[0], parts[1], nil
}

func resetIntegrationDatabase(t *testing.T) {
	t.Helper()
	requireIntegration(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const reset = `
		ALTER TABLE evaluation_state_events
			DISABLE TRIGGER trg_evaluation_events_no_truncate;
		ALTER TABLE evaluation_state
			DISABLE TRIGGER trg_evaluation_state_no_truncate;
		TRUNCATE TABLE
			evaluation_state_events,
			evaluations,
			submissions,
			team_members,
			teams,
			evaluation_state,
			users
		CASCADE;
		ALTER TABLE evaluation_state_events
			ENABLE TRIGGER trg_evaluation_events_no_truncate;
		ALTER TABLE evaluation_state
			ENABLE TRIGGER trg_evaluation_state_no_truncate;
		INSERT INTO evaluation_state (singleton_id) VALUES (1);
	`
	if _, err := integrationPool.Exec(ctx, reset); err != nil {
		t.Fatalf("reset integration database: %v", err)
	}
}

func createIntegrationUser(t *testing.T, role domain.Role) uuid.UUID {
	t.Helper()
	id := uuid.New()
	email := fmt.Sprintf("%s-%s@example.test", strings.ToLower(string(role)), id.String())
	var university, telegram any
	if role == domain.RoleUser {
		university = "Test University"
		telegram = "@test_user"
	}
	_, err := integrationPool.Exec(context.Background(), `
		INSERT INTO users (
			id, full_name, university, email, telegram, password_hash, role
		) VALUES ($1, $2, $3, $4, $5, 'integration-hash', $6)
	`, id, "Integration "+string(role), university, email, telegram, role)
	if err != nil {
		t.Fatalf("create integration %s: %v", role, err)
	}
	return id
}

func createIntegrationTeam(t *testing.T, captainID uuid.UUID) domain.Team {
	t.Helper()
	ctx := context.Background()
	tx, err := integrationPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		t.Fatalf("begin integration team transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	team := domain.Team{
		ID:         uuid.New(),
		Name:       "Team " + uuid.NewString(),
		InviteCode: strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8]),
		CaptainID:  captainID,
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO teams (id, name, invite_code, captain_id)
		VALUES ($1, $2, $3, $4)
	`, team.ID, team.Name, team.InviteCode, team.CaptainID); err != nil {
		t.Fatalf("insert integration team: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)
	`, team.ID, team.CaptainID); err != nil {
		t.Fatalf("insert integration captain membership: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit integration team: %v", err)
	}
	return team
}

func addIntegrationMember(t *testing.T, teamID, userID uuid.UUID) {
	t.Helper()
	if _, err := integrationPool.Exec(context.Background(),
		`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`,
		teamID, userID,
	); err != nil {
		t.Fatalf("add integration team member: %v", err)
	}
}

func addIntegrationSubmission(t *testing.T, teamID uuid.UUID) {
	t.Helper()
	if _, err := integrationPool.Exec(context.Background(), `
		INSERT INTO submissions (team_id, solution_url)
		VALUES ($1, 'https://example.test/solution')
	`, teamID); err != nil {
		t.Fatalf("add integration submission: %v", err)
	}
}

func integrationEvaluations(juryID, teamID uuid.UUID, score domain.Score) []domain.Evaluation {
	result := make([]domain.Evaluation, 0, domain.CriterionCount)
	for criterion := domain.FirstCriterionID; criterion <= domain.LastCriterionID; criterion++ {
		result = append(result, domain.Evaluation{
			JuryID:      juryID,
			TeamID:      teamID,
			CriterionID: criterion,
			Score:       score,
		})
	}
	return result
}

func requirePostgresCode(t *testing.T, err error, code string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("expected PostgreSQL error %s, got %v", code, err)
	}
	if postgresError.Code != code {
		t.Fatalf("expected PostgreSQL error %s, got %s: %v", code, postgresError.Code, err)
	}
}

func countIntegrationRows(t *testing.T, table string) int {
	t.Helper()
	allowed := map[string]bool{
		"users": true, "teams": true, "team_members": true, "submissions": true,
		"evaluations": true, "evaluation_state": true, "evaluation_state_events": true,
	}
	if !allowed[table] {
		t.Fatalf("unsupported integration count table %q", table)
	}
	var count int
	if err := integrationPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}
