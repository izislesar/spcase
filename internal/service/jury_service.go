package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type JuryQueryRepository interface {
	ListJuryTeams(context.Context, uuid.UUID) ([]domain.JuryTeam, error)
	EvaluationsClosed(context.Context) (bool, error)
}

// JuryWorkspace contains the team registry and its global write state.
type JuryWorkspace struct {
	Teams             []domain.JuryTeam
	EvaluationsLocked bool
}

type JuryService struct {
	queries JuryQueryRepository
}

func NewJuryService(queries JuryQueryRepository) (*JuryService, error) {
	if queries == nil {
		return nil, errors.New("jury query repository cannot be nil")
	}
	return &JuryService{queries: queries}, nil
}

func (s *JuryService) Workspace(ctx context.Context, juryID uuid.UUID) (JuryWorkspace, error) {
	if juryID == uuid.Nil {
		return JuryWorkspace{}, domain.ErrUnauthorized
	}
	teams, err := s.queries.ListJuryTeams(ctx, juryID)
	if err != nil {
		return JuryWorkspace{}, err
	}
	locked, err := s.queries.EvaluationsClosed(ctx)
	if err != nil {
		return JuryWorkspace{}, err
	}
	return JuryWorkspace{Teams: teams, EvaluationsLocked: locked}, nil
}
