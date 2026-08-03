//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"spcase.ru/backend/internal/domain"
)

func TestCanonicalSchemaIntegrity(t *testing.T) {
	resetIntegrationDatabase(t)
	ctx := context.Background()

	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	adminID := createIntegrationUser(t, domain.RoleAdmin)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	t.Run("case insensitive email uniqueness", func(t *testing.T) {
		var email string
		if err := integrationPool.QueryRow(ctx,
			`SELECT email FROM users WHERE id = $1`, captainID,
		).Scan(&email); err != nil {
			t.Fatalf("load source email: %v", err)
		}
		_, err := integrationPool.Exec(ctx, `
			INSERT INTO users (
				full_name, university, email, telegram, password_hash, role
			) VALUES ('Duplicate Email', 'University', $1, '@duplicate', 'hash', 'USER')
		`, strings.ToUpper(email))
		requirePostgresCode(t, err, "23505")
	})

	t.Run("case insensitive team name uniqueness", func(t *testing.T) {
		otherCaptainID := createIntegrationUser(t, domain.RoleUser)
		_, err := integrationPool.Exec(ctx, `
			INSERT INTO teams (name, invite_code, captain_id)
			VALUES ($1, 'UNIQUE99', $2)
		`, strings.ToUpper(team.Name), otherCaptainID)
		requirePostgresCode(t, err, "23505")
	})

	t.Run("captain membership foreign key is deferred", func(t *testing.T) {
		otherCaptainID := createIntegrationUser(t, domain.RoleUser)
		tx, err := integrationPool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin deferred FK transaction: %v", err)
		}
		defer tx.Rollback(ctx)
		_, err = tx.Exec(ctx, `
			INSERT INTO teams (name, invite_code, captain_id)
			VALUES ('Missing Captain Membership', 'MISSCAPT', $1)
		`, otherCaptainID)
		if err != nil {
			t.Fatalf("deferred FK fired before commit: %v", err)
		}
		requirePostgresCode(t, tx.Commit(ctx), "23503")
	})

	t.Run("only users may be team members", func(t *testing.T) {
		_, err := integrationPool.Exec(ctx,
			`INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`,
			team.ID, juryID,
		)
		requirePostgresCode(t, err, "23514")

		_, err = integrationPool.Exec(ctx,
			`UPDATE users SET role = 'JURY' WHERE id = $1`,
			memberID,
		)
		requirePostgresCode(t, err, "23514")
	})

	t.Run("evaluation checks and jury role are enforced", func(t *testing.T) {
		_, err := integrationPool.Exec(ctx, `
			INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
			VALUES ($1, $2, 1, 5)
		`, captainID, team.ID)
		requirePostgresCode(t, err, "23514")

		_, err = integrationPool.Exec(ctx, `
			INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
			VALUES ($1, $2, 7, 5)
		`, juryID, team.ID)
		requirePostgresCode(t, err, "23514")

		_, err = integrationPool.Exec(ctx, `
			INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
			VALUES ($1, $2, 1, 11)
		`, juryID, team.ID)
		requirePostgresCode(t, err, "23514")

		if _, err = integrationPool.Exec(ctx, `
			INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
			VALUES ($1, $2, 1, 5)
		`, juryID, team.ID); err != nil {
			t.Fatalf("insert valid jury evaluation: %v", err)
		}
		_, err = integrationPool.Exec(ctx,
			`UPDATE users SET role = 'USER', university = 'University', telegram = '@jury' WHERE id = $1`,
			juryID,
		)
		requirePostgresCode(t, err, "23514")
	})

	t.Run("evaluation state actor is admin", func(t *testing.T) {
		_, err := integrationPool.Exec(ctx, `
			UPDATE evaluation_state
			SET is_closed = TRUE, closed_at = clock_timestamp(), closed_by = $1
			WHERE singleton_id = 1
		`, juryID)
		requirePostgresCode(t, err, "23514")

		if _, err = integrationPool.Exec(ctx, `
			UPDATE evaluation_state
			SET is_closed = TRUE, closed_at = clock_timestamp(), closed_by = $1
			WHERE singleton_id = 1
		`, adminID); err != nil {
			t.Fatalf("close evaluations as admin: %v", err)
		}
		if _, err = integrationPool.Exec(ctx, `
			INSERT INTO evaluation_state_events (action, admin_id)
			VALUES ('CLOSE', $1)
		`, adminID); err != nil {
			t.Fatalf("append evaluation state event: %v", err)
		}
		_, err = integrationMigratorPool.Exec(ctx,
			`UPDATE users SET role = 'JURY' WHERE id = $1`,
			adminID,
		)
		requirePostgresCode(t, err, "23514")
	})

	t.Run("state is singleton and events are append only", func(t *testing.T) {
		_, err := integrationMigratorPool.Exec(ctx,
			`DELETE FROM evaluation_state WHERE singleton_id = 1`,
		)
		requirePostgresCode(t, err, "55000")

		_, err = integrationMigratorPool.Exec(ctx,
			`UPDATE evaluation_state_events SET action = 'OPEN'`,
		)
		requirePostgresCode(t, err, "55000")

		_, err = integrationMigratorPool.Exec(ctx,
			`DELETE FROM evaluation_state_events`,
		)
		requirePostgresCode(t, err, "55000")

		_, err = integrationMigratorPool.Exec(ctx,
			`TRUNCATE evaluation_state_events`,
		)
		requirePostgresCode(t, err, "55000")

		_, err = integrationMigratorPool.Exec(ctx,
			`TRUNCATE evaluation_state`,
		)
		requirePostgresCode(t, err, "55000")
	})
}

func TestDevelopmentSeedMigration(t *testing.T) {
	resetIntegrationDatabase(t)
	ctx := context.Background()
	if err := applySetupUpMigration(ctx, "00003_seed_dev_data.sql"); err != nil {
		t.Fatalf("apply development seed migration: %v", err)
	}

	var adminCount, juryCount, userCount int
	if err := integrationPool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE role = 'ADMIN'),
			COUNT(*) FILTER (WHERE role = 'JURY'),
			COUNT(*) FILTER (WHERE role = 'USER')
		FROM users
	`).Scan(&adminCount, &juryCount, &userCount); err != nil {
		t.Fatalf("count seeded roles: %v", err)
	}
	if adminCount != 1 || juryCount != 3 || userCount != 8 {
		t.Fatalf("unexpected seeded roles: admin=%d jury=%d user=%d", adminCount, juryCount, userCount)
	}
	if got := countIntegrationRows(t, "teams"); got != 3 {
		t.Fatalf("expected 3 seeded teams, got %d", got)
	}
	if got := countIntegrationRows(t, "team_members"); got != 6 {
		t.Fatalf("expected 6 seeded memberships, got %d", got)
	}
	if got := countIntegrationRows(t, "submissions"); got != 2 {
		t.Fatalf("expected 2 seeded submissions, got %d", got)
	}

	rows, err := integrationPool.Query(ctx, `SELECT password_hash FROM users`)
	if err != nil {
		t.Fatalf("load seeded password hashes: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var passwordHash string
		if err := rows.Scan(&passwordHash); err != nil {
			t.Fatalf("scan seeded password hash: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password")); err != nil {
			t.Fatalf("seeded password hash does not match documented dev password: %v", err)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate seeded password hashes: %v", err)
	}

	expectedIndexes := map[string]bool{
		"uq_users_email_ci":       false,
		"uq_teams_name_ci":        false,
		"idx_evaluations_team_id": false,
		"idx_teams_captain_id":    false,
	}
	indexRows, err := integrationPool.Query(ctx, `
		SELECT indexname
		FROM pg_indexes
		WHERE schemaname = current_schema()
	`)
	if err != nil {
		t.Fatalf("list schema indexes: %v", err)
	}
	defer indexRows.Close()
	for indexRows.Next() {
		var name string
		if err := indexRows.Scan(&name); err != nil {
			t.Fatalf("scan schema index: %v", err)
		}
		if _, expected := expectedIndexes[name]; expected {
			expectedIndexes[name] = true
		}
		if name == "idx_evaluations_jury_id" {
			t.Fatal("redundant jury index must not be created")
		}
	}
	for name, found := range expectedIndexes {
		if !found {
			t.Errorf("expected index %s was not created", name)
		}
	}
}

func TestDevelopmentSeedCanBeRevertedAfterUse(t *testing.T) {
	resetIntegrationDatabase(t)
	ctx := context.Background()
	if err := applySetupUpMigration(ctx, "00003_seed_dev_data.sql"); err != nil {
		t.Fatalf("apply development seed migration: %v", err)
	}
	const (
		adminID = "10000000-0000-4000-8000-000000000001"
		juryID  = "20000000-0000-4000-8000-000000000001"
		teamID  = "40000000-0000-4000-8000-000000000001"
	)
	if _, err := integrationPool.Exec(ctx, `
		INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
		VALUES ($1, $2, 1, 9);
		UPDATE evaluation_state
		SET is_closed = TRUE,
			closed_at = clock_timestamp(),
			closed_by = $3,
			updated_at = clock_timestamp()
		WHERE singleton_id = 1;
		INSERT INTO evaluation_state_events (action, admin_id)
		VALUES ('CLOSE', $3);
	`, juryID, teamID, adminID); err != nil {
		t.Fatalf("use seeded accounts before rollback: %v", err)
	}

	if err := applySetupDownMigration(ctx, "00003_seed_dev_data.sql"); err != nil {
		t.Fatalf("revert used development seed: %v", err)
	}
	if got := countIntegrationRows(t, "users"); got != 0 {
		t.Fatalf("expected seed rollback to remove users, got %d", got)
	}
	if got := countIntegrationRows(t, "teams"); got != 0 {
		t.Fatalf("expected seed rollback to remove teams, got %d", got)
	}
	if got := countIntegrationRows(t, "evaluations"); got != 0 {
		t.Fatalf("expected seed rollback to remove evaluations, got %d", got)
	}
	if got := countIntegrationRowsAsMigrator(t, "evaluation_state_events"); got != 0 {
		t.Fatalf("expected seed rollback to remove state events, got %d", got)
	}
	var closed bool
	if err := integrationPool.QueryRow(ctx,
		`SELECT is_closed FROM evaluation_state WHERE singleton_id = 1`,
	).Scan(&closed); err != nil {
		t.Fatalf("load evaluation state after seed rollback: %v", err)
	}
	if closed {
		t.Fatal("seed rollback left evaluations closed by the removed admin")
	}
}

func TestDevelopmentSeedPasswordIsBcryptCompatible(t *testing.T) {
	hash := []byte("$2a$10$4Hln1mjL0p6gjY0lTWu7Q.LmkmTPMW1KjLRKevBqE.Q5dNvYOumbu")
	if err := bcrypt.CompareHashAndPassword(hash, []byte("password")); err != nil {
		t.Fatalf("documented development password is invalid: %v", err)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte("not-password")) == nil {
		t.Fatal("seed hash unexpectedly accepts the wrong password")
	}
}

func TestSeedUUIDsAreValid(t *testing.T) {
	ids := []string{
		"10000000-0000-4000-8000-000000000001",
		"20000000-0000-4000-8000-000000000001",
		"30000000-0000-4000-8000-000000000001",
		"40000000-0000-4000-8000-000000000001",
		"50000000-0000-4000-8000-000000000001",
	}
	for _, value := range ids {
		if _, err := uuid.Parse(value); err != nil {
			t.Fatalf("invalid deterministic seed UUID %q: %v", value, err)
		}
	}
}
