package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

// ScoreUseCases contains isolated jury evaluation operations.
type ScoreUseCases interface {
	SaveEvaluations(context.Context, uuid.UUID, uuid.UUID, []service.CriterionScore) ([]domain.Evaluation, error)
	JuryEvaluations(context.Context, uuid.UUID) ([]domain.Evaluation, error)
	JuryTeamEvaluations(context.Context, uuid.UUID, uuid.UUID) ([]domain.Evaluation, error)
}

var _ ScoreUseCases = (*service.ScoreService)(nil)

// ScoreHandler serves authenticated jury evaluation endpoints.
type ScoreHandler struct {
	scores ScoreUseCases
	logger *slog.Logger
}

// NewScoreHandler creates a jury score HTTP handler.
func NewScoreHandler(scores ScoreUseCases, logger *slog.Logger) (*ScoreHandler, error) {
	if scores == nil {
		return nil, errors.New("score service cannot be nil")
	}
	return &ScoreHandler{
		scores: scores,
		logger: normalizedLogger(logger),
	}, nil
}

// SaveEvaluations atomically saves scores for the authenticated jury member.
func (h *ScoreHandler) SaveEvaluations(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleJury)
	if !ok {
		return
	}

	var input SaveEvaluationsRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	scores := make([]service.CriterionScore, 0, len(input.Scores))
	for _, score := range input.Scores {
		scores = append(scores, service.CriterionScore{
			CriterionID: score.CriterionID,
			Score:       score.Score,
		})
	}

	if _, err := h.scores.SaveEvaluations(request.Context(), principal.UserID, input.TeamID, scores); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Scores saved successfully"})
}

// Evaluations returns every score authored by the authenticated jury member.
func (h *ScoreHandler) Evaluations(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleJury)
	if !ok {
		return
	}

	evaluations, err := h.scores.JuryEvaluations(request.Context(), principal.UserID)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, evaluationsResponse(evaluations))
}

// TeamEvaluations returns the authenticated jury member's scores for teamID.
// The team ID is supplied by the router after path parsing.
func (h *ScoreHandler) TeamEvaluations(writer http.ResponseWriter, request *http.Request, teamID uuid.UUID) {
	principal, ok := requireRole(writer, request, domain.RoleJury)
	if !ok {
		return
	}

	evaluations, err := h.scores.JuryTeamEvaluations(request.Context(), principal.UserID, teamID)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, evaluationsResponse(evaluations))
}

func evaluationsResponse(evaluations []domain.Evaluation) JuryEvaluationsResponse {
	response := JuryEvaluationsResponse{
		Evaluations: make([]EvaluationResponse, 0, len(evaluations)),
	}
	for _, evaluation := range evaluations {
		response.Evaluations = append(response.Evaluations, EvaluationResponse{
			TeamID:      evaluation.TeamID,
			CriterionID: evaluation.CriterionID,
			Score:       evaluation.Score,
		})
	}
	return response
}
