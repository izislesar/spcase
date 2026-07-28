package middleware

import (
	"errors"
	"net/http"
	"strings"

	"spcase.ru/backend/internal/domain"
)

// CORSMiddleware permits credentialed browser requests only from configured
// HTTPS origins.
type CORSMiddleware struct {
	allowedOrigins map[string]struct{}
}

// NewCORSMiddleware creates restrictive credentialed CORS middleware.
func NewCORSMiddleware(origins []string) (*CORSMiddleware, error) {
	if len(origins) == 0 {
		return nil, errors.New("CORS allowed origins cannot be empty")
	}
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			return nil, errors.New("CORS allowed origin cannot be empty")
		}
		allowed[origin] = struct{}{}
	}
	return &CORSMiddleware{allowedOrigins: allowed}, nil
}

// Middleware applies CORS headers and terminates valid preflight requests.
func (m *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("CORS middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := strings.TrimSpace(request.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}
		if _, allowed := m.allowedOrigins[origin]; !allowed {
			writeDomainError(writer, http.StatusForbidden, domain.ErrForbidden)
			return
		}

		headers := writer.Header()
		headers.Add("Vary", "Origin")
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		if request.Method == http.MethodOptions {
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			headers.Set("Access-Control-Allow-Headers", "Accept, Content-Type")
			headers.Set("Access-Control-Max-Age", "600")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
