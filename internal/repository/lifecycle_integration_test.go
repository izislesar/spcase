//go:build integration

package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const (
	teamShareLockQuery   = "SELECT id FROM teams WHERE id = $1 FOR SHARE"
	teamUpdateLockQuery  = "FROM teams WHERE id = $1 FOR UPDATE"
	stateUpdateLockQuery = "SELECT is_closed FROM evaluation_state WHERE singleton_id = 1 FOR UPDATE"
)

type lifecycleTraceMarker struct{}

type blockingQueryTracer struct {
	match   string
	reached chan struct{}
	release <-chan struct{}
	once    sync.Once
}

func (tracer *blockingQueryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	if strings.Contains(strings.Join(strings.Fields(data.SQL), " "), tracer.match) {
		return context.WithValue(ctx, lifecycleTraceMarker{}, true)
	}
	return ctx
}

func (tracer *blockingQueryTracer) TraceQueryEnd(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryEndData,
) {
	if data.Err != nil || ctx.Value(lifecycleTraceMarker{}) != true {
		return
	}
	tracer.once.Do(func() {
		close(tracer.reached)
		<-tracer.release
	})
}

func TestSubmissionEvaluationCanonicalLifecycle(t *testing.T) {
	resetIntegrationDatabase(t)
	ctx := context.Background()
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberOneID := createIntegrationUser(t, domain.RoleUser)
	memberTwoID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberOneID)
	addIntegrationMember(t, team.ID, memberTwoID)
	addIntegrationSubmission(t, team.ID)

	scores, err := NewScorePostgres(integrationPool)
	if err != nil {
		t.Fatalf("create score repository: %v", err)
	}
	teams, err := NewTeamPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create team repository: %v", err)
	}
	queries, err := NewQueryPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create query repository: %v", err)
	}

	if _, err := scores.UpsertBatch(ctx, integrationEvaluations(juryID, team.ID, 7)); err != nil {
		t.Fatalf("save valid evaluations: %v", err)
	}
	if err := teams.Leave(ctx, memberTwoID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("leave while team remains eligible: %v", err)
	}
	assertLifecycleCounts(t, team.ID, 2, 1, domain.CriterionCount)

	if err := teams.Kick(ctx, captainID, memberOneID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("kick making team ineligible: %v", err)
	}
	assertLifecycleCounts(t, team.ID, 1, 0, domain.CriterionCount)
	if _, err := scores.UpsertBatch(ctx, integrationEvaluations(juryID, team.ID, 8)); !errors.Is(err, domain.ErrSubmissionNotFound) {
		t.Fatalf("score without submission error = %v, want SUBMISSION_NOT_FOUND", err)
	}
	assertLifecycleCounts(t, team.ID, 1, 0, domain.CriterionCount)

	if err := teams.Kick(ctx, captainID, uuid.New(), time.Now().Add(time.Hour)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("kick missing user error = %v, want USER_NOT_FOUND", err)
	}
	assertLifecycleCounts(t, team.ID, 1, 0, domain.CriterionCount)

	if _, err := queries.SetEvaluationClosed(ctx, createIntegrationUser(t, domain.RoleAdmin), true); err != nil {
		t.Fatalf("close evaluations: %v", err)
	}
	if _, err := queries.SetEvaluationClosed(ctx, createIntegrationUser(t, domain.RoleAdmin), false); err != nil {
		t.Fatalf("reopen evaluations: %v", err)
	}
	assertLifecycleCounts(t, team.ID, 1, 0, domain.CriterionCount)
	assertNoStructurallyInvalidLifecycleRows(t)
}

func TestScoreWriteSerializesWithTeamLifecycleMutations(t *testing.T) {
	for _, mutation := range []string{"leave", "kick", "disband"} {
		t.Run(mutation, func(t *testing.T) {
			resetIntegrationDatabase(t)
			captainID := createIntegrationUser(t, domain.RoleUser)
			memberID := createIntegrationUser(t, domain.RoleUser)
			juryID := createIntegrationUser(t, domain.RoleJury)
			team := createIntegrationTeam(t, captainID)
			addIntegrationMember(t, team.ID, memberID)
			addIntegrationSubmission(t, team.ID)

			release := make(chan struct{})
			mutationPool := newTracedIntegrationPool(t, teamUpdateLockQuery, release)
			teams, err := NewTeamPostgres(mutationPool.pool)
			if err != nil {
				t.Fatalf("create team repository: %v", err)
			}
			scores, err := NewScorePostgres(integrationPool)
			if err != nil {
				t.Fatalf("create score repository: %v", err)
			}

			mutationResult := make(chan error, 1)
			go func() {
				switch mutation {
				case "leave":
					mutationResult <- teams.Leave(context.Background(), memberID, time.Now().Add(time.Hour))
				case "kick":
					mutationResult <- teams.Kick(context.Background(), captainID, memberID, time.Now().Add(time.Hour))
				default:
					mutationResult <- teams.Disband(context.Background(), captainID, time.Now().Add(time.Hour))
				}
			}()
			awaitLifecycleBarrier(t, mutationPool.reached)

			scoreResult := make(chan error, 1)
			go func() {
				_, scoreErr := scores.UpsertBatch(context.Background(), integrationEvaluations(juryID, team.ID, 9))
				scoreResult <- scoreErr
			}()
			close(release)
			requireConcurrentResult(t, mutationResult, nil)
			requireConcurrentResult(t, scoreResult, domain.ErrSubmissionNotFound)

			if mutation == "disband" {
				if got := countIntegrationRows(t, "teams"); got != 0 {
					t.Fatalf("disband left %d teams", got)
				}
			} else {
				assertLifecycleCounts(t, team.ID, 1, 0, 0)
			}
			assertNoStructurallyInvalidLifecycleRows(t)
		})
	}
}

func TestScoreWinningRaceCommitsBeforeSubmissionInvalidation(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	release := make(chan struct{})
	scorePool := newTracedIntegrationPool(t, teamShareLockQuery, release)
	scores, _ := NewScorePostgres(scorePool.pool)
	teams, _ := NewTeamPostgres(integrationPool)
	scoreResult := make(chan error, 1)
	go func() {
		_, err := scores.UpsertBatch(context.Background(), integrationEvaluations(juryID, team.ID, 6))
		scoreResult <- err
	}()
	awaitLifecycleBarrier(t, scorePool.reached)
	leaveResult := make(chan error, 1)
	go func() { leaveResult <- teams.Leave(context.Background(), memberID, time.Now().Add(time.Hour)) }()
	close(release)
	requireConcurrentResult(t, scoreResult, nil)
	requireConcurrentResult(t, leaveResult, nil)
	assertLifecycleCounts(t, team.ID, 1, 0, domain.CriterionCount)
	assertNoStructurallyInvalidLifecycleRows(t)
}

func TestSubmissionUpsertSerializesWithEligibilityLoss(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)

	release := make(chan struct{})
	mutationPool := newTracedIntegrationPool(t, teamUpdateLockQuery, release)
	teams, _ := NewTeamPostgres(mutationPool.pool)
	submissions, _ := NewSubmissionPostgres(integrationPool)
	leaveResult := make(chan error, 1)
	go func() { leaveResult <- teams.Leave(context.Background(), memberID, time.Now().Add(time.Hour)) }()
	awaitLifecycleBarrier(t, mutationPool.reached)
	submitResult := make(chan error, 1)
	go func() {
		_, err := submissions.Upsert(context.Background(), captainID, "https://example.test/race", time.Now().Add(time.Hour))
		submitResult <- err
	}()
	close(release)
	requireConcurrentResult(t, leaveResult, nil)
	requireConcurrentResult(t, submitResult, domain.ErrMinimumTwoMembers)
	assertLifecycleCounts(t, team.ID, 1, 0, 0)
}

func TestEvaluationCloseSerializesWithScoring(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	adminID := createIntegrationUser(t, domain.RoleAdmin)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	release := make(chan struct{})
	queryPool := newTracedIntegrationPool(t, stateUpdateLockQuery, release)
	queries, _ := NewQueryPostgres(queryPool.pool)
	scores, _ := NewScorePostgres(integrationPool)
	closeResult := make(chan error, 1)
	go func() {
		_, err := queries.SetEvaluationClosed(context.Background(), adminID, true)
		closeResult <- err
	}()
	awaitLifecycleBarrier(t, queryPool.reached)
	scoreResult := make(chan error, 1)
	go func() {
		_, err := scores.UpsertBatch(context.Background(), integrationEvaluations(juryID, team.ID, 5))
		scoreResult <- err
	}()
	close(release)
	requireConcurrentResult(t, closeResult, nil)
	requireConcurrentResult(t, scoreResult, domain.ErrEvaluationLocked)
	assertLifecycleCounts(t, team.ID, 2, 1, 0)
}

func TestConcurrentScoreWritesRemainAtomicAndUnique(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	release := make(chan struct{})
	firstPool := newTracedIntegrationPool(t, teamShareLockQuery, release)
	secondPool := newTracedIntegrationPool(t, teamShareLockQuery, release)
	firstScores, _ := NewScorePostgres(firstPool.pool)
	secondScores, _ := NewScorePostgres(secondPool.pool)
	results := make(chan error, 2)
	go func() {
		_, err := firstScores.UpsertBatch(context.Background(), integrationEvaluations(juryID, team.ID, 3))
		results <- err
	}()
	go func() {
		_, err := secondScores.UpsertBatch(context.Background(), integrationEvaluations(juryID, team.ID, 8))
		results <- err
	}()
	awaitLifecycleBarrier(t, firstPool.reached)
	awaitLifecycleBarrier(t, secondPool.reached)
	close(release)
	requireConcurrentResult(t, results, nil)
	requireConcurrentResult(t, results, nil)

	rows, err := integrationPool.Query(context.Background(), `
		SELECT score FROM evaluations
		WHERE jury_id = $1 AND team_id = $2
		ORDER BY criterion_id
	`, juryID, team.ID)
	if err != nil {
		t.Fatalf("load concurrent scores: %v", err)
	}
	defer rows.Close()
	var finalScore int
	count := 0
	for rows.Next() {
		var score int
		if err := rows.Scan(&score); err != nil {
			t.Fatalf("scan concurrent score: %v", err)
		}
		if count == 0 {
			finalScore = score
		} else if score != finalScore {
			t.Fatalf("concurrent batches mixed scores: first=%d current=%d", finalScore, score)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate concurrent scores: %v", err)
	}
	if count != domain.CriterionCount || (finalScore != 3 && finalScore != 8) {
		t.Fatalf("unexpected concurrent result count=%d score=%d", count, finalScore)
	}
	assertNoStructurallyInvalidLifecycleRows(t)
}

func TestLifecycleLocksAreScopedPerTeam(t *testing.T) {
	resetIntegrationDatabase(t)
	juryID := createIntegrationUser(t, domain.RoleJury)
	firstTeam := createEligibleIntegrationTeam(t)
	secondTeam := createEligibleIntegrationTeam(t)

	release := make(chan struct{})
	blockedPool := newTracedIntegrationPool(t, teamShareLockQuery, release)
	blockedScores, _ := NewScorePostgres(blockedPool.pool)
	regularScores, _ := NewScorePostgres(integrationPool)
	firstResult := make(chan error, 1)
	go func() {
		_, err := blockedScores.UpsertBatch(context.Background(), integrationEvaluations(juryID, firstTeam.ID, 4))
		firstResult <- err
	}()
	awaitLifecycleBarrier(t, blockedPool.reached)
	secondResult := make(chan error, 1)
	go func() {
		_, err := regularScores.UpsertBatch(context.Background(), integrationEvaluations(juryID, secondTeam.ID, 7))
		secondResult <- err
	}()
	requireConcurrentResult(t, secondResult, nil)
	close(release)
	requireConcurrentResult(t, firstResult, nil)
	assertLifecycleCounts(t, firstTeam.ID, 2, 1, domain.CriterionCount)
	assertLifecycleCounts(t, secondTeam.ID, 2, 1, domain.CriterionCount)
}

type tracedIntegrationPool struct {
	pool    *pgxpool.Pool
	reached <-chan struct{}
}

func newTracedIntegrationPool(t *testing.T, match string, release <-chan struct{}) tracedIntegrationPool {
	t.Helper()
	reached := make(chan struct{})
	config := integrationPool.Config().Copy()
	config.ConnConfig.Tracer = &blockingQueryTracer{match: match, reached: reached, release: release}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("create traced integration pool: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping traced integration pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return tracedIntegrationPool{pool: pool, reached: reached}
}

func awaitLifecycleBarrier(t *testing.T, reached <-chan struct{}) {
	t.Helper()
	select {
	case <-reached:
	case <-time.After(5 * time.Second):
		t.Fatal("lifecycle operation did not reach deterministic barrier")
	}
}

func requireConcurrentResult(t *testing.T, result <-chan error, want error) {
	t.Helper()
	select {
	case err := <-result:
		if !errors.Is(err, want) {
			t.Fatalf("concurrent result = %v, want %v", err, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent lifecycle operation timed out")
	}
}

func createEligibleIntegrationTeam(t *testing.T) domain.Team {
	t.Helper()
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)
	return team
}

func assertLifecycleCounts(t *testing.T, teamID uuid.UUID, members, submissions, evaluations int) {
	t.Helper()
	var actualMembers, actualSubmissions, actualEvaluations int
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT
			(SELECT COUNT(*) FROM team_members WHERE team_id = $1),
			(SELECT COUNT(*) FROM submissions WHERE team_id = $1),
			(SELECT COUNT(*) FROM evaluations WHERE team_id = $1)
	`, teamID).Scan(&actualMembers, &actualSubmissions, &actualEvaluations); err != nil {
		t.Fatalf("load lifecycle counts: %v", err)
	}
	if actualMembers != members || actualSubmissions != submissions || actualEvaluations != evaluations {
		t.Fatalf("lifecycle counts members=%d submissions=%d evaluations=%d, want %d/%d/%d",
			actualMembers, actualSubmissions, actualEvaluations, members, submissions, evaluations)
	}
}

func assertNoStructurallyInvalidLifecycleRows(t *testing.T) {
	t.Helper()
	var orphanedTeams, ineligibleSubmissions, duplicateScores int
	if err := integrationPool.QueryRow(context.Background(), `
		SELECT
			(SELECT COUNT(*) FROM evaluations e LEFT JOIN teams t ON t.id = e.team_id WHERE t.id IS NULL),
			(SELECT COUNT(*) FROM submissions s
			 WHERE (SELECT COUNT(*) FROM team_members tm WHERE tm.team_id = s.team_id) < $1),
			(SELECT COUNT(*) FROM (
				SELECT jury_id, team_id, criterion_id
				FROM evaluations
				GROUP BY jury_id, team_id, criterion_id
				HAVING COUNT(*) > 1
			 ) duplicates)
	`, domain.MinTeamMembers).Scan(&orphanedTeams, &ineligibleSubmissions, &duplicateScores); err != nil {
		t.Fatalf("probe invalid lifecycle state: %v", err)
	}
	if orphanedTeams != 0 || ineligibleSubmissions != 0 || duplicateScores != 0 {
		t.Fatalf("invalid lifecycle rows: orphaned_team=%d ineligible_submission=%d duplicate_scores=%d",
			orphanedTeams, ineligibleSubmissions, duplicateScores)
	}
}
