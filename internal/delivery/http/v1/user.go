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

type UserUseCases interface {
	Me(context.Context, uuid.UUID) (service.UserProfile, error)
}

type UserHandler struct {
	users  UserUseCases
	logger *slog.Logger
}

func NewUserHandler(users UserUseCases, logger *slog.Logger) (*UserHandler, error) {
	if users == nil {
		return nil, errors.New("user service cannot be nil")
	}
	return &UserHandler{users: users, logger: normalizedLogger(logger)}, nil
}

func (h *UserHandler) Me(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireAuthenticated(writer, request)
	if !ok {
		return
	}
	if principal.Role != domain.RoleUser && principal.Role != domain.RoleAdmin {
		writeError(writer, http.StatusForbidden, domain.CodeForbidden, domain.ErrForbidden.Message)
		return
	}
	profile, err := h.users.Me(request.Context(), principal.UserID)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	response := UserMeResponse{
		ID: profile.User.ID, FullName: profile.User.FullName, Email: profile.User.Email,
		Role: profile.User.Role, TeamStatus: profile.MembershipState, TeamID: profile.TeamID,
	}
	if profile.User.University != nil {
		response.University = *profile.User.University
	}
	if profile.User.Telegram != nil {
		response.Telegram = *profile.User.Telegram
	}
	writeJSON(writer, http.StatusOK, response)
}
