package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

type recordingAuthenticator struct {
	validatedToken string
	claims         service.AccessTokenClaims
	projection     domain.AccountProjection
}

func (a *recordingAuthenticator) ValidateToken(token string) (service.AccessTokenClaims, error) {
	a.validatedToken = token
	return a.claims, nil
}

func (a *recordingAuthenticator) VerifyAccount(
	context.Context,
	service.AccessTokenClaims,
) (domain.AccountProjection, error) {
	return a.projection, nil
}

func TestAuthenticationAcceptsOnlyOneCookieCredential(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	newHandler := func(t *testing.T) (*recordingAuthenticator, http.Handler) {
		t.Helper()
		authenticator := &recordingAuthenticator{
			claims: service.AccessTokenClaims{
				Role: domain.RoleUser, AuthVersion: 1,
				RegisteredClaims: jwt.RegisteredClaims{Subject: userID.String()},
			},
			projection: domain.AccountProjection{
				ID: userID, Role: domain.RoleUser, AuthVersion: 1,
			},
		}
		auth, err := NewAuthMiddleware(authenticator, nil)
		if err != nil {
			t.Fatal(err)
		}
		return authenticator, auth.Middleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusNoContent)
		}))
	}

	tests := []struct {
		name       string
		prepare    func(*http.Request)
		wantStatus int
		wantToken  string
	}{
		{name: "missing", prepare: func(*http.Request) {}, wantStatus: http.StatusUnauthorized},
		{name: "bearer unsupported", prepare: func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer copied-token")
		}, wantStatus: http.StatusUnauthorized},
		{name: "query unsupported", prepare: func(request *http.Request) {
			request.URL.RawQuery = "access_token=copied-token"
		}, wantStatus: http.StatusUnauthorized},
		{name: "one cookie", prepare: func(request *http.Request) {
			request.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: "cookie-token"})
		}, wantStatus: http.StatusNoContent, wantToken: "cookie-token"},
		{name: "cookie wins over authorization header", prepare: func(request *http.Request) {
			request.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: "cookie-token"})
			request.Header.Set("Authorization", "Bearer conflicting-token")
		}, wantStatus: http.StatusNoContent, wantToken: "cookie-token"},
		{name: "duplicate cookies rejected", prepare: func(request *http.Request) {
			request.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: "first"})
			request.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: "second"})
		}, wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			authenticator, handler := newHandler(t)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
			test.prepare(request)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if authenticator.validatedToken != test.wantToken {
				t.Fatalf("validated token = %q, want %q", authenticator.validatedToken, test.wantToken)
			}
		})
	}
}

func TestRoleEscalationMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   domain.Role
		required domain.Role
	}{
		{name: "USER to JURY", actual: domain.RoleUser, required: domain.RoleJury},
		{name: "USER to ADMIN", actual: domain.RoleUser, required: domain.RoleAdmin},
		{name: "JURY to ADMIN", actual: domain.RoleJury, required: domain.RoleAdmin},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			handler := RequireRoles(test.required)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("forbidden principal reached handler")
			}))
			request := httptest.NewRequest(http.MethodPost, "/api/v1/protected", nil)
			request = request.WithContext(context.WithValue(request.Context(), authContextKey{}, Principal{
				UserID: uuid.New(), Role: test.actual,
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			assertErrorResponse(t, recorder, http.StatusForbidden, domain.CodeForbidden, domain.ErrForbidden.Message)
		})
	}
}

func TestCORSAbuseMatrix(t *testing.T) {
	t.Parallel()

	cors, err := NewCORSMiddleware([]string{"https://spcase.ru"})
	if err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})
	handler := cors.Middleware(next)

	tests := []struct {
		name       string
		origin     string
		method     string
		requested  string
		headers    string
		wantStatus int
		wantOrigin string
	}{
		{name: "no origin", method: http.MethodGet, wantStatus: http.StatusNoContent},
		{name: "allowed origin", origin: "https://spcase.ru", method: http.MethodGet, wantStatus: http.StatusNoContent, wantOrigin: "https://spcase.ru"},
		{name: "untrusted origin", origin: "https://evil.example", method: http.MethodGet, wantStatus: http.StatusForbidden},
		{name: "null origin", origin: "null", method: http.MethodPost, wantStatus: http.StatusForbidden},
		{name: "allowed preflight", origin: "https://spcase.ru", method: http.MethodOptions, requested: http.MethodPost, headers: "content-type", wantStatus: http.StatusNoContent, wantOrigin: "https://spcase.ru"},
		{name: "forged preflight method", origin: "https://spcase.ru", method: http.MethodOptions, requested: http.MethodPatch, headers: "content-type", wantStatus: http.StatusForbidden},
		{name: "forged preflight header", origin: "https://spcase.ru", method: http.MethodOptions, requested: http.MethodPost, headers: "authorization", wantStatus: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(test.method, "/api/v1/team/create", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			if test.requested != "" {
				request.Header.Set("Access-Control-Request-Method", test.requested)
			}
			if test.headers != "" {
				request.Header.Set("Access-Control-Request-Headers", test.headers)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if actual := recorder.Header().Get("Access-Control-Allow-Origin"); actual != test.wantOrigin {
				t.Fatalf("allow origin = %q, want %q", actual, test.wantOrigin)
			}
		})
	}
}

func TestRecoveryDoesNotExposePanicValue(t *testing.T) {
	t.Parallel()

	const secret = "secret-password-jury-key-token"
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := NewRecoveryMiddleware(logger).Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(secret)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/test", nil))

	assertErrorResponse(t, recorder, http.StatusInternalServerError, domain.CodeInternal, domain.ErrInternal.Message)
	if strings.Contains(recorder.Body.String(), secret) || strings.Contains(logs.String(), secret) {
		t.Fatalf("panic secret exposed; response=%q logs=%q", recorder.Body.String(), logs.String())
	}
	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("decode recovery log: %v", err)
	}
	if entry["panic_type"] != "string" {
		t.Fatalf("panic_type = %#v, want string", entry["panic_type"])
	}
}
