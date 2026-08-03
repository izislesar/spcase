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

const (
	integrationMigratorDatabaseEnvironment = "SPCASE_TEST_MIGRATOR_DATABASE_URL"
	integrationAppDatabaseEnvironment      = "SPCASE_TEST_APP_DATABASE_URL"
	integrationLegacyDatabaseEnvironment   = "SPCASE_TEST_DATABASE_URL"
)

var (
	integrationMigratorPool *pgxpool.Pool
	integrationPool         *pgxpool.Pool
	integrationEnabled      bool
	integrationSchema       string
)

func TestMain(m *testing.M) {
	migratorURL := os.Getenv(integrationMigratorDatabaseEnvironment)
	appURL := os.Getenv(integrationAppDatabaseEnvironment)
	if migratorURL == "" && appURL == "" {
		if os.Getenv(integrationLegacyDatabaseEnvironment) != "" {
			fatalIntegrationSetup("configure test database credentials",
				fmt.Errorf("%s is no longer accepted; set %s and %s",
					integrationLegacyDatabaseEnvironment,
					integrationMigratorDatabaseEnvironment,
					integrationAppDatabaseEnvironment,
				))
		}
		os.Exit(m.Run())
	}
	if migratorURL == "" {
		fatalIntegrationSetup("configure test database credentials",
			fmt.Errorf("%s is required when %s is set",
				integrationMigratorDatabaseEnvironment, integrationAppDatabaseEnvironment))
	}
	if appURL == "" {
		fatalIntegrationSetup("configure test database credentials",
			fmt.Errorf("%s is required when %s is set",
				integrationAppDatabaseEnvironment, integrationMigratorDatabaseEnvironment))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	migratorConfig, err := pgxpool.ParseConfig(migratorURL)
	if err != nil {
		fatalIntegrationSetup("parse migrator test database URL", err)
	}
	migratorPool, err := pgxpool.NewWithConfig(ctx, migratorConfig)
	if err != nil {
		fatalIntegrationSetup("connect with migrator test credentials", err)
	}
	if err := verifyIntegrationRole(ctx, migratorPool, "spcase_migrator"); err != nil {
		migratorPool.Close()
		fatalIntegrationSetup("verify migrator test connection", err)
	}
	if err := verifyProductionMigrationState(ctx, migratorPool); err != nil {
		migratorPool.Close()
		fatalIntegrationSetup("verify production migration state", err)
	}

	integrationSchema = "spcase_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quotedSchema := pgx.Identifier{integrationSchema}.Sanitize()
	if _, err := migratorPool.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		migratorPool.Close()
		fatalIntegrationSetup("create isolated test schema", err)
	}

	integrationMigratorPool, err = newIntegrationSchemaPool(ctx, migratorURL, quotedSchema, 4)
	if err != nil {
		_ = dropIntegrationSchema(ctx, migratorPool, quotedSchema)
		migratorPool.Close()
		fatalIntegrationSetup("connect migrator to isolated test schema", err)
	}
	if err := applySetupUpMigration(ctx, "00001_init_schema.sql"); err != nil {
		cleanupFailedIntegrationSetup(ctx, migratorPool, quotedSchema)
		fatalIntegrationSetup("apply isolated schema migration", err)
	}
	if err := applySetupUpMigration(ctx, "00002_add_indexes.sql"); err != nil {
		cleanupFailedIntegrationSetup(ctx, migratorPool, quotedSchema)
		fatalIntegrationSetup("apply isolated index migration", err)
	}
	if err := copyRuntimePrivileges(ctx, migratorPool, quotedSchema); err != nil {
		cleanupFailedIntegrationSetup(ctx, migratorPool, quotedSchema)
		fatalIntegrationSetup("grant isolated runtime privileges", err)
	}

	integrationPool, err = newIntegrationSchemaPool(ctx, appURL, quotedSchema, 12)
	if err != nil {
		cleanupFailedIntegrationSetup(ctx, migratorPool, quotedSchema)
		fatalIntegrationSetup("connect application to isolated test schema", err)
	}
	if err := verifyRuntimeIntegrationPool(ctx); err != nil {
		integrationPool.Close()
		cleanupFailedIntegrationSetup(ctx, migratorPool, quotedSchema)
		fatalIntegrationSetup("verify runtime repository connection", err)
	}
	integrationEnabled = true

	exitCode := m.Run()
	integrationPool.Close()
	integrationMigratorPool.Close()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := dropIntegrationSchema(cleanupCtx, migratorPool, quotedSchema); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "database integration cleanup: %v\n", err)
		if exitCode == 0 {
			exitCode = 1
		}
	}
	cleanupCancel()
	migratorPool.Close()
	os.Exit(exitCode)
}

func fatalIntegrationSetup(operation string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "database integration setup: %s: %v\n", operation, err)
	os.Exit(1)
}

func newIntegrationSchemaPool(
	ctx context.Context,
	databaseURL string,
	quotedSchema string,
	maxConnections int32,
) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	config.ConnConfig.RuntimeParams["search_path"] = quotedSchema + ",public"
	config.MaxConns = maxConnections
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func verifyIntegrationRole(ctx context.Context, pool *pgxpool.Pool, expected string) error {
	var role string
	if err := pool.QueryRow(ctx, `SELECT current_user`).Scan(&role); err != nil {
		return err
	}
	if role != expected {
		return fmt.Errorf("connected as role %q, expected %q", role, expected)
	}
	return nil
}

func verifyProductionMigrationState(ctx context.Context, pool *pgxpool.Pool) error {
	var version int64
	if err := pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version_id) FILTER (WHERE is_applied), 0)
		FROM public.goose_db_version
	`).Scan(&version); err != nil {
		return fmt.Errorf("read Goose migration version as spcase_migrator: %w", err)
	}
	if version != 5 {
		return fmt.Errorf("production migration version is %d, expected 5", version)
	}
	return nil
}

func verifyRuntimeIntegrationPool(ctx context.Context) error {
	var role, schema, usersOwner string
	if err := integrationPool.QueryRow(ctx, `
		SELECT current_user, current_schema(), pg_get_userbyid(c.relowner)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = current_schema() AND c.relname = 'users'
	`).Scan(&role, &schema, &usersOwner); err != nil {
		return err
	}
	if role != "spcase_app" {
		return fmt.Errorf("repository pool connected as role %q, expected %q", role, "spcase_app")
	}
	if schema != integrationSchema {
		return fmt.Errorf("repository pool resolved schema %q, expected isolated schema", schema)
	}
	if usersOwner != "spcase_migrator" {
		return fmt.Errorf("isolated application objects owned by %q, expected %q", usersOwner, "spcase_migrator")
	}
	return nil
}

func copyRuntimePrivileges(ctx context.Context, sourcePool *pgxpool.Pool, quotedSchema string) error {
	if _, err := sourcePool.Exec(ctx, "GRANT USAGE ON SCHEMA "+quotedSchema+" TO spcase_app"); err != nil {
		return err
	}

	rows, err := sourcePool.Query(ctx, `
		SELECT table_name, privilege_type
		FROM information_schema.role_table_grants
		WHERE table_schema = 'public' AND grantee = 'spcase_app'
		ORDER BY table_name, privilege_type
	`)
	if err != nil {
		return err
	}
	type tablePrivilege struct{ table, privilege string }
	var grants []tablePrivilege
	for rows.Next() {
		var grant tablePrivilege
		if err := rows.Scan(&grant.table, &grant.privilege); err != nil {
			rows.Close()
			return err
		}
		switch grant.privilege {
		case "SELECT", "INSERT", "UPDATE", "DELETE":
		default:
			rows.Close()
			return fmt.Errorf("refuse to copy unexpected runtime privilege %q", grant.privilege)
		}
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(grants) == 0 {
		return errors.New("production runtime table privileges are missing")
	}
	for _, grant := range grants {
		table := pgx.Identifier{integrationSchema, grant.table}.Sanitize()
		if _, err := sourcePool.Exec(ctx,
			"GRANT "+grant.privilege+" ON TABLE "+table+" TO spcase_app",
		); err != nil {
			return fmt.Errorf("copy %s privilege for %s: %w", grant.privilege, grant.table, err)
		}
	}
	typeName := pgx.Identifier{integrationSchema, "user_role"}.Sanitize()
	if _, err := sourcePool.Exec(ctx, "GRANT USAGE ON TYPE "+typeName+" TO spcase_app"); err != nil {
		return err
	}
	return nil
}

func cleanupFailedIntegrationSetup(ctx context.Context, migratorPool *pgxpool.Pool, quotedSchema string) {
	if integrationPool != nil {
		integrationPool.Close()
	}
	if integrationMigratorPool != nil {
		integrationMigratorPool.Close()
	}
	_ = dropIntegrationSchema(ctx, migratorPool, quotedSchema)
	migratorPool.Close()
}

func dropIntegrationSchema(ctx context.Context, pool *pgxpool.Pool, quotedSchema string) error {
	_, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE")
	return err
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if !integrationEnabled {
		t.Skipf("set %s and %s to run PostgreSQL integration tests",
			integrationMigratorDatabaseEnvironment, integrationAppDatabaseEnvironment)
	}
}

func applySetupUpMigration(ctx context.Context, name string) error {
	up, _, err := readMigrationSections(name)
	if err != nil {
		return err
	}
	if _, err := integrationMigratorPool.Exec(ctx, up); err != nil {
		return fmt.Errorf("execute %s Up section: %w", name, err)
	}
	return nil
}

func applySetupDownMigration(ctx context.Context, name string) error {
	_, down, err := readMigrationSections(name)
	if err != nil {
		return err
	}
	if _, err := integrationMigratorPool.Exec(ctx, down); err != nil {
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
	if _, err := integrationMigratorPool.Exec(ctx, reset); err != nil {
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

func countIntegrationRowsAsMigrator(t *testing.T, table string) int {
	t.Helper()
	if table != "evaluation_state_events" {
		t.Fatalf("migrator count is not permitted for %q", table)
	}
	var count int
	if err := integrationMigratorPool.QueryRow(
		context.Background(), "SELECT COUNT(*) FROM "+table,
	).Scan(&count); err != nil {
		t.Fatalf("count %s as migrator: %v", table, err)
	}
	return count
}
