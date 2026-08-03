package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

const readinessTimeout = 2 * time.Second

type PublicUseCases interface {
	Info() service.PublicInfo
	Schedule() []service.ScheduleEvent
	FAQ() []service.FAQItem
	NoTeam() (string, string)
}

type DatabasePinger interface {
	Ping(context.Context) error
}

type PublicHandler struct {
	public PublicUseCases
	db     DatabasePinger
	logger *slog.Logger
	now    func() time.Time
}

func NewPublicHandler(public PublicUseCases, db DatabasePinger, logger *slog.Logger) (*PublicHandler, error) {
	if public == nil || db == nil {
		return nil, errors.New("public dependencies cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &PublicHandler{public: public, db: db, logger: logger, now: time.Now}, nil
}

func (h *PublicHandler) Live(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, HealthResponse{Status: "ok", Timestamp: h.now().UTC()})
}

func (h *PublicHandler) Ready(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), readinessTimeout)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		h.logger.WarnContext(
			request.Context(),
			"database readiness check failed",
			slog.String("event", "database_readiness_failed"),
			slog.String("request_id", middleware.RequestIDFromContext(request.Context())),
			slog.String("error", err.Error()),
		)
		writeError(writer, http.StatusServiceUnavailable, domain.CodeNotReady, domain.ErrNotReady.Message)
		return
	}
	writeJSON(writer, http.StatusOK, HealthResponse{Status: "ready", Timestamp: h.now().UTC()})
}

func (h *PublicHandler) Info(writer http.ResponseWriter, _ *http.Request) {
	info := h.public.Info()
	writeJSON(writer, http.StatusOK, ChampionshipInfoResponse{
		RegistrationDeadline: info.RegistrationDeadline,
		SubmissionDeadline:   info.SubmissionDeadline,
		IsRegistrationOpen:   info.RegistrationOpen,
		IsSubmissionOpen:     info.SubmissionOpen,
	})
}

func (h *PublicHandler) Schedule(writer http.ResponseWriter, _ *http.Request) {
	events := h.public.Schedule()
	response := ScheduleResponse{Events: make([]ScheduleEventResponse, 0, len(events))}
	for _, event := range events {
		response.Events = append(response.Events, ScheduleEventResponse(event))
	}
	writeJSON(writer, http.StatusOK, response)
}

func (h *PublicHandler) FAQ(writer http.ResponseWriter, _ *http.Request) {
	items := h.public.FAQ()
	response := FAQResponse{FAQ: make([]FAQItemResponse, 0, len(items))}
	for _, item := range items {
		response.FAQ = append(response.FAQ, FAQItemResponse(item))
	}
	writeJSON(writer, http.StatusOK, response)
}

func (h *PublicHandler) NoTeam(writer http.ResponseWriter, _ *http.Request) {
	message, telegramURL := h.public.NoTeam()
	writeJSON(writer, http.StatusOK, NoTeamResponse{
		Message: message, TelegramChatURL: telegramURL,
	})
}
