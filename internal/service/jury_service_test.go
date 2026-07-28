package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type fakeJuryQueries struct {
	teams  []domain.JuryTeam
	locked bool
	err    error
}

func (f fakeJuryQueries) ListJuryTeams(context.Context, uuid.UUID) ([]domain.JuryTeam, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.teams, nil
}

func (f fakeJuryQueries) EvaluationsClosed(context.Context) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.locked, nil
}

func TestJuryWorkspaceIncludesEvaluationLock(t *testing.T) {
	team := domain.JuryTeam{TeamID: uuid.New(), TeamName: "Team"}
	service, err := NewJuryService(fakeJuryQueries{
		teams:  []domain.JuryTeam{team},
		locked: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	workspace, err := service.Workspace(context.Background(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if !workspace.EvaluationsLocked {
		t.Fatal("expected evaluations to be locked")
	}
	if len(workspace.Teams) != 1 || workspace.Teams[0].TeamID != team.TeamID {
		t.Fatalf("unexpected teams: %#v", workspace.Teams)
	}
}

func TestJuryWorkspaceRejectsMissingPrincipal(t *testing.T) {
	service, err := NewJuryService(fakeJuryQueries{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Workspace(context.Background(), uuid.Nil); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}
