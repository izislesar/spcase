package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

type fakeAuthenticator struct {
	claims     service.AccessTokenClaims
	projection domain.AccountProjection
}

func (f fakeAuthenticator) ValidateToken(string) (service.AccessTokenClaims, error) {
	return f.claims, nil
}

func (f fakeAuthenticator) VerifyAccount(
	context.Context,
	service.AccessTokenClaims,
) (domain.AccountProjection, error) {
	return f.projection, nil
}

func TestAuthMiddlewareUsesCurrentAccountProjection(t *testing.T) {
	userID := uuid.New()
	auth, err := NewAuthMiddleware(fakeAuthenticator{
		claims: service.AccessTokenClaims{
			Role: domain.RoleUser, AuthVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{Subject: userID.String()},
		},
		projection: domain.AccountProjection{
			ID: userID, Role: domain.RoleAdmin, AuthVersion: 2,
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := auth.Middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		principal, ok := PrincipalFromContext(request.Context())
		if !ok || principal.Role != domain.RoleAdmin {
			t.Fatalf("principal = %#v, %v", principal, ok)
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: "token"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestRoleMiddlewareRejectsDifferentRole(t *testing.T) {
	next := RequireRoles(domain.RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(context.WithValue(request.Context(), authContextKey{}, Principal{
		UserID: uuid.New(), Role: domain.RoleUser,
	}))
	recorder := httptest.NewRecorder()
	next.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "FORBIDDEN") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestCORSMiddlewareAllowsOnlyConfiguredOrigin(t *testing.T) {
	cors, err := NewCORSMiddleware([]string{"https://spcase.ru"})
	if err != nil {
		t.Fatal(err)
	}
	handler := cors.Middleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://evil.example")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d", recorder.Code)
	}
}

type lockedMutations bool

func (l lockedMutations) MutationsLocked() bool { return bool(l) }

func TestHardLockMiddleware(t *testing.T) {
	lock, err := NewHardLockMiddleware(lockedMutations(true))
	if err != nil {
		t.Fatal(err)
	}
	handler := lock.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d", recorder.Code)
	}
}
