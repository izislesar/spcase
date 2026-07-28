package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMembershipStateFor(t *testing.T) {
	userID := uuid.New()
	teamID := uuid.New()
	membership := &TeamMembership{TeamID: teamID, UserID: userID}

	tests := []struct {
		name       string
		membership *TeamMembership
		captainID  uuid.UUID
		want       MembershipState
	}{
		{name: "without team", want: MembershipNoTeam},
		{name: "member", membership: membership, captainID: uuid.New(), want: MembershipInTeam},
		{name: "captain", membership: membership, captainID: userID, want: MembershipCaptain},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := MembershipStateFor(userID, test.membership, test.captainID); got != test.want {
				t.Fatalf("MembershipStateFor() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestStatusBadge(t *testing.T) {
	if got := StatusBadge(1, false); got != TeamStatusSearching {
		t.Fatalf("one member = %q", got)
	}
	if got := StatusBadge(2, false); got != TeamStatusReady {
		t.Fatalf("two members = %q", got)
	}
	if got := StatusBadge(1, true); got != TeamStatusSubmitted {
		t.Fatalf("submission = %q", got)
	}
}

func TestJuryEvaluationTotal(t *testing.T) {
	juryID, teamID := uuid.New(), uuid.New()
	evaluations := make([]Evaluation, 0, CriterionCount)
	for criterion := FirstCriterionID; criterion <= LastCriterionID; criterion++ {
		evaluations = append(evaluations, Evaluation{
			JuryID: juryID, TeamID: teamID, CriterionID: criterion, Score: MaxScore,
		})
	}

	total, err := JuryEvaluationTotal(evaluations)
	if err != nil {
		t.Fatalf("JuryEvaluationTotal() error = %v", err)
	}
	if total != MaximumJuryTotal {
		t.Fatalf("JuryEvaluationTotal() = %d, want %d", total, MaximumJuryTotal)
	}

	evaluations[5].CriterionID = FirstCriterionID
	if _, err := JuryEvaluationTotal(evaluations); !errors.Is(err, ErrInvalidEvaluation) {
		t.Fatalf("duplicate criterion error = %v", err)
	}
}
