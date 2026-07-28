package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type AdminRepository interface {
	AdminStats(context.Context) (domain.AdminStats, error)
	SetEvaluationClosed(context.Context, uuid.UUID, bool) (domain.EvaluationState, error)
}

type AdminService struct {
	repository AdminRepository
}

func NewAdminService(repository AdminRepository) (*AdminService, error) {
	if repository == nil {
		return nil, errors.New("admin repository cannot be nil")
	}
	return &AdminService{repository: repository}, nil
}

func (s *AdminService) Stats(ctx context.Context) (domain.AdminStats, error) {
	return s.repository.AdminStats(ctx)
}

func (s *AdminService) CloseEvaluations(
	ctx context.Context,
	adminID uuid.UUID,
) (domain.EvaluationState, error) {
	if adminID == uuid.Nil {
		return domain.EvaluationState{}, domain.ErrUnauthorized
	}
	return s.repository.SetEvaluationClosed(ctx, adminID, true)
}

func (s *AdminService) OpenEvaluations(
	ctx context.Context,
	adminID uuid.UUID,
) (domain.EvaluationState, error) {
	if adminID == uuid.Nil {
		return domain.EvaluationState{}, domain.ErrUnauthorized
	}
	return s.repository.SetEvaluationClosed(ctx, adminID, false)
}
