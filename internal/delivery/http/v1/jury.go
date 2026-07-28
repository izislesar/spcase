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

type JuryUseCases interface {
	Workspace(context.Context, uuid.UUID) (service.JuryWorkspace, error)
}

type JuryHandler struct {
	jury   JuryUseCases
	logger *slog.Logger
}

func NewJuryHandler(jury JuryUseCases, logger *slog.Logger) (*JuryHandler, error) {
	if jury == nil {
		return nil, errors.New("jury service cannot be nil")
	}
	return &JuryHandler{jury: jury, logger: normalizedLogger(logger)}, nil
}

func (h *JuryHandler) Teams(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleJury)
	if !ok {
		return
	}
	workspace, err := h.jury.Workspace(request.Context(), principal.UserID)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	response := JuryTeamsResponse{
		Teams:             make([]JuryTeamResponse, 0, len(workspace.Teams)),
		EvaluationsLocked: workspace.EvaluationsLocked,
	}
	for _, team := range workspace.Teams {
		response.Teams = append(response.Teams, JuryTeamResponse{
			TeamID: team.TeamID, TeamName: team.TeamName, SolutionURL: team.SolutionURL,
			IsEvaluatedByMe: team.EvaluatedByMe, MembersCount: team.MembersCount,
		})
	}
	writeJSON(writer, http.StatusOK, response)
}
