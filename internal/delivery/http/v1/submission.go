package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type SubmissionUseCases interface {
	Submit(context.Context, uuid.UUID, string) (domain.Submission, error)
}

type SubmissionHandler struct {
	submissions SubmissionUseCases
	logger      *slog.Logger
}

func NewSubmissionHandler(
	submissions SubmissionUseCases,
	logger *slog.Logger,
) (*SubmissionHandler, error) {
	if submissions == nil {
		return nil, errors.New("submission service cannot be nil")
	}
	return &SubmissionHandler{submissions: submissions, logger: normalizedLogger(logger)}, nil
}

func (h *SubmissionHandler) Submit(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}
	var input SubmitSolutionRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	solutionURL, err := domain.NormalizeSolutionURL(input.SolutionURL)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	submission, err := h.submissions.Submit(
		request.Context(), principal.UserID, solutionURL,
	)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, SubmitSolutionResponse{
		Status: "submitted", SolutionURL: submission.SolutionURL, UpdatedAt: submission.UpdatedAt,
	})
}
