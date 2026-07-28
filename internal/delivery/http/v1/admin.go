package v1

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type AdminUseCases interface {
	Stats(context.Context) (domain.AdminStats, error)
	CloseEvaluations(context.Context, uuid.UUID) (domain.EvaluationState, error)
	OpenEvaluations(context.Context, uuid.UUID) (domain.EvaluationState, error)
}

type ExportUseCases interface {
	WriteXLSX(context.Context, io.Writer) error
}

type AdminHandler struct {
	admin  AdminUseCases
	export ExportUseCases
	logger *slog.Logger
}

func NewAdminHandler(
	admin AdminUseCases,
	export ExportUseCases,
	logger *slog.Logger,
) (*AdminHandler, error) {
	if admin == nil || export == nil {
		return nil, errors.New("admin services cannot be nil")
	}
	return &AdminHandler{admin: admin, export: export, logger: normalizedLogger(logger)}, nil
}

func (h *AdminHandler) Stats(writer http.ResponseWriter, request *http.Request) {
	if _, ok := requireRole(writer, request, domain.RoleAdmin); !ok {
		return
	}
	stats, err := h.admin.Stats(request.Context())
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, AdminStatsResponse{
		TotalUsers: stats.TotalUsers, TotalTeams: stats.TotalTeams,
		SubmittedSolutions: stats.SubmittedSolutions, TotalJuries: stats.TotalJuries,
		EvaluationsClosed: stats.EvaluationsClosed,
	})
}

func (h *AdminHandler) CloseEvaluations(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleAdmin)
	if !ok {
		return
	}
	if _, err := h.admin.CloseEvaluations(request.Context(), principal.UserID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Evaluations closed"})
}

func (h *AdminHandler) OpenEvaluations(writer http.ResponseWriter, request *http.Request) {
	principal, ok := requireRole(writer, request, domain.RoleAdmin)
	if !ok {
		return
	}
	if _, err := h.admin.OpenEvaluations(request.Context(), principal.UserID); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	writeJSON(writer, http.StatusOK, MessageResponse{Message: "Evaluations opened"})
}

func (h *AdminHandler) ExportExcel(writer http.ResponseWriter, request *http.Request) {
	if _, ok := requireRole(writer, request, domain.RoleAdmin); !ok {
		return
	}

	var output bytes.Buffer
	if err := h.export.WriteXLSX(request.Context(), &output); err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	writer.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	writer.Header().Set("Content-Disposition", `attachment; filename="hackathon_results.xlsx"`)
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(output.Bytes()); err != nil {
		logInternalError(h.logger, request, err)
	}
}
