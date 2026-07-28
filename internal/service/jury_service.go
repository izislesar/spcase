package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type JuryQueryRepository interface {
	ListJuryTeams(context.Context, uuid.UUID) ([]domain.JuryTeam, error)
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

func (s *JuryService) Teams(ctx context.Context, juryID uuid.UUID) ([]domain.JuryTeam, error) {
	if juryID == uuid.Nil {
		return nil, domain.ErrUnauthorized
	}
	return s.queries.ListJuryTeams(ctx, juryID)
}
