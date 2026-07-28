//go:build integration

package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"spcase.ru/backend/internal/domain"
)

func TestDisbandCascadesRelationsButKeepsAccounts(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)
	if _, err := integrationPool.Exec(context.Background(), `
		INSERT INTO evaluations (jury_id, team_id, criterion_id, score)
		VALUES ($1, $2, 1, 8)
	`, juryID, team.ID); err != nil {
		t.Fatalf("insert cascade evaluation: %v", err)
	}

	repository, err := NewTeamPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create team repository: %v", err)
	}
	if err := repository.Disband(context.Background(), captainID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("disband team: %v", err)
	}

	if got := countIntegrationRows(t, "teams"); got != 0 {
		t.Fatalf("expected no teams after disband, got %d", got)
	}
	if got := countIntegrationRows(t, "team_members"); got != 0 {
		t.Fatalf("expected memberships to cascade, got %d", got)
	}
	if got := countIntegrationRows(t, "submissions"); got != 0 {
		t.Fatalf("expected submission to cascade, got %d", got)
	}
	if got := countIntegrationRows(t, "evaluations"); got != 0 {
		t.Fatalf("expected evaluations to cascade, got %d", got)
	}
	if got := countIntegrationRows(t, "users"); got != 3 {
		t.Fatalf("expected all 3 accounts to remain, got %d", got)
	}
}

func TestConcurrentJoinWithOneFreePlace(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberOneID := createIntegrationUser(t, domain.RoleUser)
	memberTwoID := createIntegrationUser(t, domain.RoleUser)
	contenderOneID := createIntegrationUser(t, domain.RoleUser)
	contenderTwoID := createIntegrationUser(t, domain.RoleUser)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberOneID)
	addIntegrationMember(t, team.ID, memberTwoID)

	repository, err := NewTeamPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create team repository: %v", err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, contenderID := range []uuid.UUID{contenderOneID, contenderTwoID} {
		go func(userID uuid.UUID) {
			<-start
			_, joinErr := repository.Join(context.Background(), userID, team.InviteCode)
			results <- joinErr
		}(contenderID)
	}
	close(start)

	var successCount, teamFullCount int
	for range 2 {
		select {
		case joinErr := <-results:
			switch {
			case joinErr == nil:
				successCount++
			case errors.Is(joinErr, domain.ErrTeamFull):
				teamFullCount++
			default:
				t.Fatalf("unexpected concurrent join result: %v", joinErr)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent join timed out")
		}
	}
	if successCount != 1 || teamFullCount != 1 {
		t.Fatalf("expected one success and one TEAM_FULL, got success=%d full=%d", successCount, teamFullCount)
	}
	var memberCount int
	if err := integrationPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM team_members WHERE team_id = $1`, team.ID,
	).Scan(&memberCount); err != nil {
		t.Fatalf("count joined members: %v", err)
	}
	if memberCount != domain.MaxTeamMembers {
		t.Fatalf("expected %d members, got %d", domain.MaxTeamMembers, memberCount)
	}
}

func TestMutationLockOrderCompletesWithoutDeadlock(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberOneID := createIntegrationUser(t, domain.RoleUser)
	memberTwoID := createIntegrationUser(t, domain.RoleUser)
	joinerID := createIntegrationUser(t, domain.RoleUser)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberOneID)
	addIntegrationMember(t, team.ID, memberTwoID)

	repository, err := NewTeamPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create team repository: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	lockAt := time.Now().Add(time.Hour)
	start := make(chan struct{})
	results := make(chan error, 5)
	operations := []func() error{
		func() error {
			_, operationErr := repository.Join(ctx, joinerID, team.InviteCode)
			return operationErr
		},
		func() error { return repository.Leave(ctx, memberOneID, lockAt) },
		func() error { return repository.Kick(ctx, captainID, memberTwoID, lockAt) },
		func() error { return repository.TransferOwnership(ctx, captainID, memberOneID, lockAt) },
		func() error { return repository.Disband(ctx, captainID, lockAt) },
	}
	var workers sync.WaitGroup
	for _, operation := range operations {
		workers.Add(1)
		go func(run func() error) {
			defer workers.Done()
			<-start
			results <- run()
		}(operation)
	}
	close(start)
	workers.Wait()
	close(results)

	for operationErr := range results {
		if errors.Is(operationErr, context.DeadlineExceeded) {
			t.Fatalf("mutation lock order timed out: %v", operationErr)
		}
		var postgresError *pgconn.PgError
		if errors.As(operationErr, &postgresError) && postgresError.Code == "40P01" {
			t.Fatalf("mutation lock order caused a deadlock: %v", operationErr)
		}
	}
}

func TestHardLockIsRecheckedAfterDatabaseRowLock(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)

	blocker, err := integrationPool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin blocker transaction: %v", err)
	}
	defer blocker.Rollback(context.Background())
	if _, err := blocker.Exec(context.Background(),
		`SELECT id FROM teams WHERE id = $1 FOR UPDATE`, team.ID,
	); err != nil {
		t.Fatalf("lock team row in blocker transaction: %v", err)
	}

	repository, err := NewTeamPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create team repository: %v", err)
	}
	lockAt := time.Now().Add(250 * time.Millisecond)
	result := make(chan error, 1)
	go func() {
		result <- repository.Leave(context.Background(), memberID, lockAt)
	}()

	time.Sleep(350 * time.Millisecond)
	if err := blocker.Commit(context.Background()); err != nil {
		t.Fatalf("release blocker transaction: %v", err)
	}
	select {
	case mutationErr := <-result:
		if !errors.Is(mutationErr, domain.ErrMutationsLocked) {
			t.Fatalf("expected mutation lock after crossing deadline, got %v", mutationErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("hard-lock boundary test timed out")
	}

	var exists bool
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2
		)
	`, team.ID, memberID).Scan(&exists); err != nil {
		t.Fatalf("check member after hard lock: %v", err)
	}
	if !exists {
		t.Fatal("member was removed despite transaction-level hard lock")
	}
}

func TestSubmissionDeletionWhenTeamBecomesUndersizedIsAtomic(t *testing.T) {
	for _, mutation := range []string{"leave", "kick"} {
		t.Run(mutation, func(t *testing.T) {
			resetIntegrationDatabase(t)
			captainID := createIntegrationUser(t, domain.RoleUser)
			memberID := createIntegrationUser(t, domain.RoleUser)
			team := createIntegrationTeam(t, captainID)
			addIntegrationMember(t, team.ID, memberID)
			addIntegrationSubmission(t, team.ID)

			repository, err := NewTeamPostgres(integrationPool)
			if err != nil {
				t.Fatalf("create team repository: %v", err)
			}
			lockAt := time.Now().Add(time.Hour)
			switch mutation {
			case "leave":
				err = repository.Leave(context.Background(), memberID, lockAt)
			case "kick":
				err = repository.Kick(context.Background(), captainID, memberID, lockAt)
			}
			if err != nil {
				t.Fatalf("%s member: %v", mutation, err)
			}

			var memberCount, submissionCount int
			if err := integrationPool.QueryRow(context.Background(), `
				SELECT
					(SELECT COUNT(*) FROM team_members WHERE team_id = $1),
					(SELECT COUNT(*) FROM submissions WHERE team_id = $1)
			`, team.ID).Scan(&memberCount, &submissionCount); err != nil {
				t.Fatalf("load post-mutation state: %v", err)
			}
			if memberCount != 1 || submissionCount != 0 {
				t.Fatalf("expected one member and no submission, got members=%d submissions=%d",
					memberCount, submissionCount)
			}
		})
	}
}
