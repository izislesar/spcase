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

// TeamUseCases contains participant team operations used by TeamHandler.
type TeamUseCases interface {
	Create(context.Context, uuid.UUID, string) (domain.Team, error)
	Join(context.Context, uuid.UUID, string) (domain.Team, error)
	MyTeam(context.Context, uuid.UUID) (service.TeamDetails, error)
	Leave(context.Context, uuid.UUID) error
	Kick(context.Context, uuid.UUID, uuid.UUID) error
	TransferOwnership(context.Context, uuid.UUID, uuid.UUID) error
	Disband(context.Context, uuid.UUID) error
}

var _ TeamUseCases = (*service.TeamService)(nil)

// TeamHandler serves authenticated participant team endpoints.
type TeamHandler struct {
	teams  TeamUseCases
	logger *slog.Logger
}

// NewTeamHandler creates a team HTTP handler.
func NewTeamHandler(teams TeamUseCases, logger *slog.Logger) (*TeamHandler, error) {
	if teams == nil {
		return nil, errors.New("team service cannot be nil")
	}
	return &TeamHandler{
		teams:  teams,
		logger: normalizedLogger(logger),
	}, nil
}

// Create creates a team owned by the authenticated participant.
func (h *TeamHandler) Create(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}

	var input CreateTeamRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	team, err := h.teams.Create(request.Context(), principal.UserID, input.Name)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	writeJSON(writer, http.StatusCreated, CreateTeamResponse{
		ID:         team.ID,
		Name:       team.Name,
		InviteCode: team.InviteCode,
		CaptainID:  team.CaptainID,
	})
}

// Join adds the authenticated participant to a team by invite code.
func (h *TeamHandler) Join(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}

	var input JoinTeamRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	team, err := h.teams.Join(request.Context(), principal.UserID, input.InviteCode)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	writeJSON(writer, http.StatusOK, JoinTeamResponse{
		Message:  "Successfully joined team",
		TeamID:   team.ID,
		TeamName: team.Name,
	})
}

// My returns the authenticated participant's team and roster.
func (h *TeamHandler) My(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}

	details, err := h.teams.MyTeam(request.Context(), principal.UserID)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	members := make([]TeamMemberResponse, 0, len(details.Members))
	for _, member := range details.Members {
		members = append(members, TeamMemberResponse{
			ID:         member.ID,
			FullName:   member.FullName,
			University: member.University,
			Telegram:   member.Telegram,
			IsCaptain:  member.IsCaptain,
		})
	}
	var submission *TeamSubmissionResponse
	if details.Submission != nil {
		submission = &TeamSubmissionResponse{
			SolutionURL: details.Submission.SolutionURL,
			UpdatedAt:   details.Submission.UpdatedAt,
		}
	}
	writeJSON(writer, http.StatusOK, MyTeamResponse{
		ID:              details.Team.ID,
		Name:            details.Team.Name,
		InviteCode:      details.Team.InviteCode,
		CaptainID:       details.Team.CaptainID,
		StatusBadge:     details.Status,
		MutationsLocked: details.MutationsLocked,
		Members:         members,
		Submission:      submission,
	})
}

// Leave removes the authenticated regular member from their team.
func (h *TeamHandler) Leave(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}
	if err := h.teams.Leave(request.Context(), principal.UserID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Successfully left team"})
}

// Kick removes a member from the authenticated captain's team.
func (h *TeamHandler) Kick(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}

	var input KickTeamMemberRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	if err := h.teams.Kick(request.Context(), principal.UserID, input.UserID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "User kicked successfully"})
}

// TransferOwnership transfers captaincy to another current member.
func (h *TeamHandler) TransferOwnership(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}

	var input TransferOwnershipRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	if err := h.teams.TransferOwnership(request.Context(), principal.UserID, input.NewCaptainID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Ownership transferred successfully"})
}

// Disband deletes the authenticated captain's team.
func (h *TeamHandler) Disband(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleUser)
	if !ok {
		return
	}
	if err := h.teams.Disband(request.Context(), principal.UserID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Team disbanded successfully"})
}
