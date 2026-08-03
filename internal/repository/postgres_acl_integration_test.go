//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	aclMigratorDatabaseEnvironment = "SPCASE_TEST_MIGRATOR_DATABASE_URL"
	aclAppDatabaseEnvironment      = "SPCASE_TEST_APP_DATABASE_URL"
)

func TestPostgresRolePrivileges(t *testing.T) {
	migratorURL := os.Getenv(aclMigratorDatabaseEnvironment)
	appURL := os.Getenv(aclAppDatabaseEnvironment)
	if migratorURL == "" || appURL == "" {
		t.Skipf("set %s and %s to run PostgreSQL ACL tests", aclMigratorDatabaseEnvironment, aclAppDatabaseEnvironment)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	migratorPool := openACLTestPool(t, ctx, migratorURL)
	defer migratorPool.Close()
	appPool := openACLTestPool(t, ctx, appURL)
	defer appPool.Close()

	t.Run("migrator owns and can maintain Goose metadata", func(t *testing.T) {
		var tableOwner, sequenceOwner string
		if err := migratorPool.QueryRow(ctx, `
			SELECT pg_get_userbyid(c.relowner)
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND c.relname = 'goose_db_version'
		`).Scan(&tableOwner); err != nil {
			t.Fatalf("read Goose table owner: %v", err)
		}
		if err := migratorPool.QueryRow(ctx, `
			SELECT pg_get_userbyid(c.relowner)
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND c.relname = 'goose_db_version_id_seq'
		`).Scan(&sequenceOwner); err != nil {
			t.Fatalf("read Goose sequence owner: %v", err)
		}
		if tableOwner != "spcase_migrator" || sequenceOwner != "spcase_migrator" {
			t.Fatalf("unexpected Goose owners: table=%s sequence=%s", tableOwner, sequenceOwner)
		}

		var isolationMigrationApplied bool
		if err := migratorPool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM public.goose_db_version
				WHERE version_id = 5 AND is_applied
			)
		`).Scan(&isolationMigrationApplied); err != nil {
			t.Fatalf("migrator reads Goose metadata: %v", err)
		}
		if !isolationMigrationApplied {
			t.Fatal("Goose metadata isolation migration 00005 is not applied")
		}
		if _, err := migratorPool.Exec(ctx, `UPDATE public.goose_db_version SET is_applied = is_applied WHERE false`); err != nil {
			t.Fatalf("migrator updates Goose metadata: %v", err)
		}
	})

	t.Run("application runtime DML and defaults remain available", func(t *testing.T) {
		assertRuntimeTablePrivileges(t, ctx, appPool)
		testApplicationDML(t, ctx, appPool)
	})

	t.Run("application cannot access Goose metadata", func(t *testing.T) {
		deniedStatements := []string{
			`SELECT * FROM public.goose_db_version`,
			`INSERT INTO public.goose_db_version (version_id, is_applied) VALUES (-1, false)`,
			`UPDATE public.goose_db_version SET is_applied = is_applied WHERE false`,
			`DELETE FROM public.goose_db_version WHERE false`,
			`TRUNCATE TABLE public.goose_db_version`,
			`SELECT nextval('public.goose_db_version_id_seq')`,
			`SELECT last_value FROM public.goose_db_version_id_seq`,
			`SELECT setval('public.goose_db_version_id_seq', 1, false)`,
		}
		for _, statement := range deniedStatements {
			_, err := appPool.Exec(ctx, statement)
			requirePostgresCode(t, err, "42501")
		}
	})

	t.Run("application DDL and ownership changes remain denied", func(t *testing.T) {
		deniedStatements := []string{
			`CREATE TABLE public.spcase_acl_probe (id integer)`,
			`ALTER TABLE public.users ADD COLUMN spcase_acl_probe integer`,
			`DROP TABLE public.users`,
			`ALTER TABLE public.users OWNER TO spcase_app`,
		}
		for _, statement := range deniedStatements {
			_, err := appPool.Exec(ctx, statement)
			requirePostgresCode(t, err, "42501")
		}
	})
}

func openACLTestPool(t *testing.T, ctx context.Context, databaseURL string) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse PostgreSQL ACL test URL: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect PostgreSQL ACL test pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping PostgreSQL ACL test pool: %v", err)
	}
	return pool
}

func assertRuntimeTablePrivileges(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	expected := map[string]string{
		"users":                   "SELECT,INSERT,UPDATE",
		"teams":                   "SELECT,INSERT,UPDATE,DELETE",
		"team_members":            "SELECT,INSERT,DELETE",
		"submissions":             "SELECT,INSERT,UPDATE,DELETE",
		"evaluations":             "SELECT,INSERT,UPDATE",
		"evaluation_state":        "SELECT,UPDATE",
		"evaluation_state_events": "INSERT",
	}
	for table, privileges := range expected {
		var allowed bool
		if err := pool.QueryRow(ctx, `SELECT has_table_privilege(current_user, $1, $2)`, "public."+table, privileges).Scan(&allowed); err != nil {
			t.Fatalf("check %s privileges: %v", table, err)
		}
		if !allowed {
			t.Errorf("runtime privileges missing on %s: %s", table, privileges)
		}
	}
}

func testApplicationDML(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin application ACL transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	insertUser := func(role, suffix string) uuid.UUID {
		var id uuid.UUID
		university := any(nil)
		telegram := any(nil)
		if role == "USER" {
			university = "ACL University"
			telegram = "@acl_" + suffix
		}
		err := tx.QueryRow(ctx, `
			INSERT INTO users (full_name, university, email, telegram, password_hash, role)
			VALUES ($1, $2, $3, $4, 'acl-hash', $5)
			RETURNING id
		`, "ACL "+role, university, "acl-"+suffix+"@example.test", telegram, role).Scan(&id)
		if err != nil {
			t.Fatalf("insert %s with database defaults: %v", role, err)
		}
		return id
	}

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	userID := insertUser("USER", suffix+"-user")
	juryID := insertUser("JURY", suffix+"-jury")
	adminID := insertUser("ADMIN", suffix+"-admin")

	var teamID uuid.UUID
	inviteCode := strings.ToUpper(suffix[:8])
	if err := tx.QueryRow(ctx, `
		INSERT INTO teams (name, invite_code, captain_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, "ACL Team "+suffix, inviteCode, userID).Scan(&teamID); err != nil {
		t.Fatalf("insert team with database defaults: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, teamID, userID); err != nil {
		t.Fatalf("insert team membership: %v", err)
	}

	var submissionID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO submissions (team_id, solution_url)
		VALUES ($1, 'https://example.test/acl')
		RETURNING id
	`, teamID).Scan(&submissionID); err != nil {
		t.Fatalf("insert submission with database defaults: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
		VALUES ($1, $2, 1, 5)
	`, juryID, teamID); err != nil {
		t.Fatalf("insert evaluation with database defaults: %v", err)
	}

	updates := []struct {
		name string
		sql  string
		args []any
	}{
		{name: "user", sql: `UPDATE users SET full_name = full_name WHERE id = $1`, args: []any{userID}},
		{name: "team", sql: `UPDATE teams SET updated_at = clock_timestamp() WHERE id = $1`, args: []any{teamID}},
		{name: "submission", sql: `UPDATE submissions SET solution_url = solution_url WHERE id = $1`, args: []any{submissionID}},
		{name: "evaluation", sql: `UPDATE evaluations SET score = 6 WHERE jury_id = $1 AND team_id = $2`, args: []any{juryID, teamID}},
		{name: "evaluation state", sql: `UPDATE evaluation_state SET is_closed = true, closed_at = clock_timestamp(), closed_by = $1 WHERE singleton_id = 1`, args: []any{adminID}},
	}
	for _, update := range updates {
		if _, err := tx.Exec(ctx, update.sql, update.args...); err != nil {
			t.Fatalf("update %s: %v", update.name, err)
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO evaluation_state_events (action, admin_id) VALUES ('CLOSE', $1)`, adminID); err != nil {
		t.Fatalf("insert evaluation state event: %v", err)
	}

	for _, table := range []string{"users", "teams", "team_members", "submissions", "evaluations", "evaluation_state"} {
		var count int
		if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM public.%s", table)).Scan(&count); err != nil {
			t.Fatalf("select %s: %v", table, err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM submissions WHERE id = $1`, submissionID); err != nil {
		t.Fatalf("delete submission: %v", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID); err != nil {
		t.Fatalf("delete team membership: %v", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM teams WHERE id = $1`, teamID); err != nil {
		t.Fatalf("delete team: %v", err)
	}
}
