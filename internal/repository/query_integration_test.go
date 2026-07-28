//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

func TestEvaluationBatchUpsertIsAtomicAndComplete(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	repository, err := NewScorePostgres(integrationPool)
	if err != nil {
		t.Fatalf("create score repository: %v", err)
	}
	if _, err := integrationPool.Exec(context.Background(), `
		ALTER TABLE evaluations
		ADD CONSTRAINT integration_reject_criterion_four
		CHECK (criterion_id <> 4)
	`); err != nil {
		t.Fatalf("add integration failure constraint: %v", err)
	}
	if _, err := repository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryID, team.ID, 7),
	); err == nil {
		t.Fatal("expected forced batch failure")
	}
	if got := countIntegrationRows(t, "evaluations"); got != 0 {
		t.Fatalf("failed batch persisted %d partial evaluation rows", got)
	}
	if _, err := integrationPool.Exec(context.Background(), `
		ALTER TABLE evaluations DROP CONSTRAINT integration_reject_criterion_four
	`); err != nil {
		t.Fatalf("drop integration failure constraint: %v", err)
	}

	saved, err := repository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryID, team.ID, 7),
	)
	if err != nil {
		t.Fatalf("save complete evaluation batch: %v", err)
	}
	if len(saved) != domain.CriterionCount {
		t.Fatalf("expected %d saved evaluations, got %d", domain.CriterionCount, len(saved))
	}
	if got := countIntegrationRows(t, "evaluations"); got != domain.CriterionCount {
		t.Fatalf("expected %d evaluation rows, got %d", domain.CriterionCount, got)
	}

	if _, err := repository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryID, team.ID, domain.MaxScore),
	); err != nil {
		t.Fatalf("replace complete evaluation batch: %v", err)
	}
	total, err := repository.TeamTotal(context.Background(), team.ID)
	if err != nil {
		t.Fatalf("load team total: %v", err)
	}
	if total.Total != domain.MaximumJuryTotal || total.EvaluatedByCount != 1 {
		t.Fatalf("unexpected replaced total: %+v", total)
	}
	if got := countIntegrationRows(t, "evaluations"); got != domain.CriterionCount {
		t.Fatalf("upsert duplicated evaluation rows: got %d", got)
	}
}

func TestExportAggregationDoesNotMultiplyScoresByMembers(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryOneID := createIntegrationUser(t, domain.RoleJury)
	juryTwoID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	scoreRepository, err := NewScorePostgres(integrationPool)
	if err != nil {
		t.Fatalf("create score repository: %v", err)
	}
	if _, err := scoreRepository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryOneID, team.ID, 1),
	); err != nil {
		t.Fatalf("save first jury scores: %v", err)
	}
	if _, err := scoreRepository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryTwoID, team.ID, 2),
	); err != nil {
		t.Fatalf("save second jury scores: %v", err)
	}

	queryRepository, err := NewQueryPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create query repository: %v", err)
	}
	summary, err := queryRepository.ExportSummary(context.Background())
	if err != nil {
		t.Fatalf("load export summary: %v", err)
	}
	if len(summary) != 1 {
		t.Fatalf("expected one export row, got %d", len(summary))
	}
	const expectedTotal = 18
	if summary[0].TotalScore != expectedTotal {
		t.Fatalf("expected unmultiplied score %d, got %d", expectedTotal, summary[0].TotalScore)
	}
	if summary[0].TotalMembers != 2 || summary[0].EvaluatedByCount != 2 {
		t.Fatalf("unexpected export aggregation: %+v", summary[0])
	}
	details, err := queryRepository.ExportDetails(context.Background())
	if err != nil {
		t.Fatalf("load export details: %v", err)
	}
	if len(details) != 2*domain.CriterionCount {
		t.Fatalf("expected %d detail rows, got %d", 2*domain.CriterionCount, len(details))
	}
}

func TestCriticalQueriesHaveExecutablePlans(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)
	scoreRepository, err := NewScorePostgres(integrationPool)
	if err != nil {
		t.Fatalf("create score repository: %v", err)
	}
	if _, err := scoreRepository.UpsertBatch(
		context.Background(),
		integrationEvaluations(juryID, team.ID, 5),
	); err != nil {
		t.Fatalf("seed query-plan scores: %v", err)
	}

	tests := []struct {
		name      string
		query     string
		arguments []any
	}{
		{
			name:      "roster",
			query:     listTeamMembersQuery,
			arguments: []any{team.ID},
		},
		{
			name:      "jury list",
			query:     listJuryTeamsQuery,
			arguments: []any{juryID, domain.CriterionCount},
		},
		{
			name:  "team totals",
			query: listTeamTotalsQuery,
		},
		{
			name:  "xlsx summary",
			query: exportSummaryQuery,
		},
		{
			name:  "xlsx details",
			query: exportDetailsQuery,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var plan string
			explainQuery := "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) " + test.query
			if err := integrationPool.QueryRow(
				context.Background(),
				explainQuery,
				test.arguments...,
			).Scan(&plan); err != nil {
				t.Fatalf("EXPLAIN ANALYZE %s: %v", test.name, err)
			}
			if !strings.Contains(plan, `"Execution Time"`) {
				t.Fatalf("EXPLAIN ANALYZE %s returned no execution metrics: %s", test.name, plan)
			}
		})
	}
}

func TestJuryListUsesPersonalEvaluationCoverage(t *testing.T) {
	resetIntegrationDatabase(t)
	captainID := createIntegrationUser(t, domain.RoleUser)
	memberID := createIntegrationUser(t, domain.RoleUser)
	juryID := createIntegrationUser(t, domain.RoleJury)
	otherJuryID := createIntegrationUser(t, domain.RoleJury)
	team := createIntegrationTeam(t, captainID)
	addIntegrationMember(t, team.ID, memberID)
	addIntegrationSubmission(t, team.ID)

	scoreRepository, err := NewScorePostgres(integrationPool)
	if err != nil {
		t.Fatalf("create score repository: %v", err)
	}
	if _, err := scoreRepository.UpsertBatch(
		context.Background(),
		integrationEvaluations(otherJuryID, team.ID, 9),
	); err != nil {
		t.Fatalf("save other jury scores: %v", err)
	}

	queryRepository, err := NewQueryPostgres(integrationPool)
	if err != nil {
		t.Fatalf("create query repository: %v", err)
	}
	teams, err := queryRepository.ListJuryTeams(context.Background(), juryID)
	if err != nil {
		t.Fatalf("list teams for jury: %v", err)
	}
	if len(teams) != 1 {
		t.Fatalf("expected one submitted team, got %d", len(teams))
	}
	if teams[0].EvaluatedByMe {
		t.Fatal("other jury scores leaked into personal evaluation coverage")
	}
	if teams[0].MembersCount != 2 {
		t.Fatalf("expected two members, got %d", teams[0].MembersCount)
	}
}

func TestIntegrationEvaluationIDsAreDistinct(t *testing.T) {
	juryID := uuid.New()
	teamID := uuid.New()
	evaluations := integrationEvaluations(juryID, teamID, 1)
	criteria := make(map[domain.CriterionID]struct{}, len(evaluations))
	for _, evaluation := range evaluations {
		criteria[evaluation.CriterionID] = struct{}{}
	}
	if len(criteria) != domain.CriterionCount {
		t.Fatalf("integration fixture has %d criteria", len(criteria))
	}
}
