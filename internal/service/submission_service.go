package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type SubmissionRepository interface {
	Upsert(context.Context, uuid.UUID, string, time.Time) (domain.Submission, error)
}

type SubmissionService struct {
	submissions SubmissionRepository
	deadline    time.Time
	now         func() time.Time
}

func NewSubmissionService(
	submissions SubmissionRepository,
	deadline time.Time,
) (*SubmissionService, error) {
	if submissions == nil {
		return nil, errors.New("submission repository cannot be nil")
	}
	if deadline.IsZero() {
		return nil, errors.New("submission deadline cannot be zero")
	}
	return &SubmissionService{submissions: submissions, deadline: deadline.UTC(), now: time.Now}, nil
}

func (s *SubmissionService) Submit(
	ctx context.Context,
	captainID uuid.UUID,
	rawURL string,
) (domain.Submission, error) {
	if !s.now().UTC().Before(s.deadline) {
		return domain.Submission{}, domain.ErrDeadlinePassed
	}
	solutionURL, err := domain.NormalizeSolutionURL(rawURL)
	if err != nil {
		return domain.Submission{}, err
	}
	return s.submissions.Upsert(ctx, captainID, solutionURL, s.deadline)
}
