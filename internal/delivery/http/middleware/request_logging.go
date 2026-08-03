package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// unmatchedRoute is logged when the router did not match a registered pattern.
// Raw URL paths are never logged here to keep log cardinality bounded; the
// Nginx access log remains the source for raw paths.
const unmatchedRoute = "unmatched"

// RequestLogging emits one structured http_request_completed event per HTTP
// request with bounded-cardinality fields.
type RequestLogging struct {
	logger *slog.Logger
}

// NewRequestLogging creates request logging middleware.
func NewRequestLogging(logger *slog.Logger) *RequestLogging {
	if logger == nil {
		logger = slog.Default()
	}
	return &RequestLogging{logger: logger}
}

// Middleware logs method, route pattern, status, duration, and request ID for
// every request. 5xx responses are logged at error level, all others at info.
func (m *RequestLogging) Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("request logging middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}

		next.ServeHTTP(recorder, request)

		route := request.Pattern
		if route == "" {
			route = unmatchedRoute
		}
		attributes := []any{
			slog.String("event", "http_request_completed"),
			slog.String("request_id", RequestIDFromContext(request.Context())),
			slog.String("method", request.Method),
			slog.String("route", route),
			slog.Int("status", recorder.status),
			slog.Int64("duration_ms", time.Since(started).Milliseconds()),
		}
		if recorder.status >= http.StatusInternalServerError {
			m.logger.ErrorContext(request.Context(), "HTTP request completed", attributes...)
			return
		}
		m.logger.InfoContext(request.Context(), "HTTP request completed", attributes...)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (recorder *statusRecorder) WriteHeader(status int) {
	if recorder.wroteHeader {
		return
	}
	recorder.wroteHeader = true
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *statusRecorder) Write(payload []byte) (int, error) {
	if !recorder.wroteHeader {
		recorder.WriteHeader(http.StatusOK)
	}
	return recorder.ResponseWriter.Write(payload)
}

// Unwrap lets http.ResponseController retain access to the original response
// writer and its optional interfaces.
func (recorder *statusRecorder) Unwrap() http.ResponseWriter {
	return recorder.ResponseWriter
}
