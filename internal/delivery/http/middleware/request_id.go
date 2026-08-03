package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeaderName is the header used to propagate request correlation IDs.
const RequestIDHeaderName = "X-Request-ID"

const (
	maxInboundRequestIDLength = 64
	minInboundRequestIDLength = 8
)

type requestIDContextKey struct{}

// RequestID assigns a correlation ID to every request. A valid inbound
// X-Request-ID is accepted; anything missing, malformed, oversized, or
// containing characters outside [A-Za-z0-9._-] is replaced with a new UUID.
// The ID is returned in a response header and stored in the request context.
// It is a correlation token only and must never be treated as authentication.
func RequestID(next http.Handler) http.Handler {
	if next == nil {
		panic("request ID middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get(RequestIDHeaderName)
		if !isValidInboundRequestID(requestID) {
			requestID = uuid.NewString()
		}

		writer.Header().Set(RequestIDHeaderName, requestID)
		ctx := context.WithValue(request.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// RequestIDFromContext returns the correlation ID attached by RequestID.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	if !ok {
		return ""
	}
	return requestID
}

func isValidInboundRequestID(requestID string) bool {
	if len(requestID) < minInboundRequestIDLength || len(requestID) > maxInboundRequestIDLength {
		return false
	}
	for _, character := range requestID {
		switch {
		case character >= 'a' && character <= 'z',
			character >= 'A' && character <= 'Z',
			character >= '0' && character <= '9',
			character == '-', character == '.', character == '_':
		default:
			return false
		}
	}
	return true
}
