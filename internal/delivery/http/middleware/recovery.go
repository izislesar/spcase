package middleware

import (
	"log/slog"
	"net/http"

	"spcase.ru/backend/internal/domain"
)

// RecoveryMiddleware converts handler panics into opaque JSON errors.
type RecoveryMiddleware struct {
	logger *slog.Logger
}

// NewRecoveryMiddleware creates panic recovery middleware.
func NewRecoveryMiddleware(logger *slog.Logger) *RecoveryMiddleware {
	if logger == nil {
		logger = slog.Default()
	}
	return &RecoveryMiddleware{logger: logger}
}

// Middleware recovers panics without exposing their details to clients.
func (m *RecoveryMiddleware) Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("recovery middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				m.logger.ErrorContext(
					request.Context(),
					"HTTP handler panic",
					slog.String("method", request.Method),
					slog.String("path", request.URL.Path),
					slog.Any("panic", recovered),
				)
				writeDomainError(writer, http.StatusInternalServerError, domain.ErrInternal)
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
