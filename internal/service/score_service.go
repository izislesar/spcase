package service

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type ScoreRepository interface {
	UpsertBatch(context.Context, []domain.Evaluation) ([]domain.Evaluation, error)
	ListByJuryID(context.Context, uuid.UUID) ([]domain.Evaluation, error)
	ListByJuryAndTeamID(context.Context, uuid.UUID, uuid.UUID) ([]domain.Evaluation, error)
	TeamTotal(context.Context, uuid.UUID) (domain.TeamScoreTotal, error)
}

type CriterionScore struct {
	CriterionID domain.CriterionID
	Score       domain.Score
}

type ScoreService struct {
	scores ScoreRepository
}

func NewScoreService(scores ScoreRepository) (*ScoreService, error) {
	if scores == nil {
		return nil, errors.New("score repository cannot be nil")
	}
	return &ScoreService{scores: scores}, nil
}

func (s *ScoreService) SaveEvaluations(
	ctx context.Context,
	juryID, teamID uuid.UUID,
	scores []CriterionScore,
) ([]domain.Evaluation, error) {
	if juryID == uuid.Nil || teamID == uuid.Nil || len(scores) != domain.CriterionCount {
		return nil, domain.ErrInvalidEvaluation
	}
	evaluations := make([]domain.Evaluation, 0, domain.CriterionCount)
	for _, criterionScore := range scores {
		evaluations = append(evaluations, domain.Evaluation{
			JuryID: juryID, TeamID: teamID,
			CriterionID: criterionScore.CriterionID, Score: criterionScore.Score,
		})
	}
	if _, err := domain.JuryEvaluationTotal(evaluations); err != nil {
		return nil, err
	}
	sort.Slice(evaluations, func(i, j int) bool {
		return evaluations[i].CriterionID < evaluations[j].CriterionID
	})
	return s.scores.UpsertBatch(ctx, evaluations)
}

func (s *ScoreService) JuryEvaluations(
	ctx context.Context,
	juryID uuid.UUID,
) ([]domain.Evaluation, error) {
	if juryID == uuid.Nil {
		return nil, domain.ErrInvalidEvaluation
	}
	return s.scores.ListByJuryID(ctx, juryID)
}

func (s *ScoreService) JuryTeamEvaluations(
	ctx context.Context,
	juryID, teamID uuid.UUID,
) ([]domain.Evaluation, error) {
	if juryID == uuid.Nil || teamID == uuid.Nil {
		return nil, domain.ErrInvalidEvaluation
	}
	return s.scores.ListByJuryAndTeamID(ctx, juryID, teamID)
}

func (s *ScoreService) TeamTotal(
	ctx context.Context,
	teamID uuid.UUID,
) (domain.TeamScoreTotal, error) {
	if teamID == uuid.Nil {
		return domain.TeamScoreTotal{}, domain.ErrTeamNotFound
	}
	return s.scores.TeamTotal(ctx, teamID)
}
