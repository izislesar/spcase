// Package middleware contains HTTP middleware for authentication and access control.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

const AccessTokenCookieName = "access_token"

type authContextKey struct{}

type Authenticator interface {
	ValidateToken(string) (service.AccessTokenClaims, error)
	VerifyAccount(context.Context, service.AccessTokenClaims) (domain.AccountProjection, error)
}

type Principal struct {
	UserID uuid.UUID
	Role   domain.Role
}

type AuthMiddleware struct {
	authenticator Authenticator
	logger        *slog.Logger
}

func NewAuthMiddleware(
	authenticator Authenticator,
	logger *slog.Logger,
) (*AuthMiddleware, error) {
	if authenticator == nil {
		return nil, errors.New("authenticator cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthMiddleware{authenticator: authenticator, logger: logger}, nil
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("auth middleware next handler cannot be nil")
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cookies := request.CookiesNamed(AccessTokenCookieName)
		if len(cookies) != 1 || strings.TrimSpace(cookies[0].Value) == "" {
			writeUnauthorized(writer)
			return
		}
		claims, err := m.authenticator.ValidateToken(cookies[0].Value)
		if err != nil {
			writeUnauthorized(writer)
			return
		}
		projection, err := m.authenticator.VerifyAccount(request.Context(), claims)
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrAccountDisabled):
				writeDomainError(writer, http.StatusUnauthorized, domain.ErrAccountDisabled)
			case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, domain.ErrUserNotFound):
				writeUnauthorized(writer)
			default:
				m.logger.ErrorContext(
					request.Context(),
					"authentication account lookup failed",
					slog.String("method", request.Method),
					slog.String("path", request.URL.Path),
					slog.String("error", err.Error()),
				)
				writeDomainError(writer, http.StatusInternalServerError, domain.ErrInternal)
			}
			return
		}
		principal := Principal{UserID: projection.ID, Role: projection.Role}
		ctx := context.WithValue(request.Context(), authContextKey{}, principal)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	principal, ok := ctx.Value(authContextKey{}).(Principal)
	if !ok || principal.UserID == uuid.Nil || !principal.Role.IsValid() {
		return Principal{}, false
	}
	return principal, true
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	principal, ok := PrincipalFromContext(ctx)
	return principal.UserID, ok
}

func RoleFromContext(ctx context.Context) (domain.Role, bool) {
	principal, ok := PrincipalFromContext(ctx)
	return principal.Role, ok
}

func writeUnauthorized(writer http.ResponseWriter) {
	writeDomainError(writer, http.StatusUnauthorized, domain.ErrUnauthorized)
}
