package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	httpmiddleware "spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

const (
	revocationTestEmail       = "revocation@example.test"
	revocationTestPassword    = "initial-password"
	revocationTestNewPassword = "rotated-password"
	revocationTestJWTSecret   = "revocation-test-jwt-secret-that-is-long-enough"
)

type revocableAuthUsers struct {
	mu      sync.RWMutex
	user    domain.User
	deleted bool
}

func newRevocableAuthUsers(t *testing.T) *revocableAuthUsers {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(revocationTestPassword),
		bcrypt.MinCost,
	)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	return &revocableAuthUsers{user: domain.User{
		ID:           uuid.New(),
		FullName:     "Revocation Test User",
		Email:        revocationTestEmail,
		PasswordHash: string(passwordHash),
		Role:         domain.RoleUser,
		AuthVersion:  1,
		CreatedAt:    time.Now().UTC(),
	}}
}

func (r *revocableAuthUsers) Create(
	context.Context,
	domain.User,
) (domain.User, error) {
	return domain.User{}, errors.New("create is not supported by revocation test repository")
}

func (r *revocableAuthUsers) GetByEmail(
	_ context.Context,
	email string,
) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.deleted || !strings.EqualFold(email, r.user.Email) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return r.user, nil
}

func (r *revocableAuthUsers) GetAccountProjection(
	_ context.Context,
	id uuid.UUID,
) (domain.AccountProjection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.deleted || id != r.user.ID {
		return domain.AccountProjection{}, domain.ErrUserNotFound
	}
	return domain.AccountProjection{
		ID:          r.user.ID,
		Role:        r.user.Role,
		AuthVersion: r.user.AuthVersion,
		DisabledAt:  r.user.DisabledAt,
	}, nil
}

func (r *revocableAuthUsers) incrementAuthVersion() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.user.AuthVersion++
}

func (r *revocableAuthUsers) setDisabled(disabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if disabled {
		now := time.Now().UTC()
		r.user.DisabledAt = &now
	} else {
		r.user.DisabledAt = nil
	}
	r.user.AuthVersion++
}

func (r *revocableAuthUsers) changePassword(t *testing.T, password string) {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash replacement password: %v", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.user.PasswordHash = string(passwordHash)
	r.user.AuthVersion++
}

func (r *revocableAuthUsers) changeRole(role domain.Role) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.user.Role = role
}

func (r *revocableAuthUsers) delete() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleted = true
}

type revocationTestApplication struct {
	authHandler *AuthHandler
	auth        *httpmiddleware.AuthMiddleware
	users       *revocableAuthUsers
}

func newRevocationTestApplication(t *testing.T) revocationTestApplication {
	t.Helper()

	users := newRevocableAuthUsers(t)
	authService, err := service.NewAuthService(
		users,
		revocationTestJWTSecret,
		"revocation-test-jury-secret",
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}
	authHandler, err := NewAuthHandler(authService, "spcase.ru", nil)
	if err != nil {
		t.Fatalf("create auth handler: %v", err)
	}
	auth, err := httpmiddleware.NewAuthMiddleware(authService, nil)
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}
	return revocationTestApplication{
		authHandler: authHandler,
		auth:        auth,
		users:       users,
	}
}

func (a revocationTestApplication) login(
	t *testing.T,
	password string,
) (*httptest.ResponseRecorder, *http.Cookie) {
	t.Helper()

	body, err := json.Marshal(LoginRequest{
		Email:    revocationTestEmail,
		Password: password,
	})
	if err != nil {
		t.Fatalf("marshal login request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	a.authHandler.Login(recorder, request)

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == httpmiddleware.AccessTokenCookieName {
			return recorder, cookie
		}
	}
	return recorder, nil
}

func (a revocationTestApplication) protected(
	roles ...domain.Role,
) http.Handler {
	return a.auth.Middleware(
		httpmiddleware.RequireRoles(roles...)(http.HandlerFunc(
			func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNoContent)
			},
		)),
	)
}

func requestWithToken(handler http.Handler, cookie *http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertRevocationError(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	status int,
	code domain.ErrorCode,
) {
	t.Helper()

	if recorder.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, status, recorder.Body.String())
	}
	var response ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v; body = %s", err, recorder.Body.String())
	}
	if response.Error.Code != code {
		t.Fatalf("error code = %q, want %q", response.Error.Code, code)
	}
}

func TestJWTAuthVersionRevocationAndRelogin(t *testing.T) {
	app := newRevocationTestApplication(t)
	login, originalCookie := app.login(t, revocationTestPassword)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d; body = %s", login.Code, login.Body.String())
	}
	if originalCookie == nil {
		t.Fatal("login did not set access_token cookie")
	}
	if !originalCookie.HttpOnly || !originalCookie.Secure {
		t.Fatalf("cookie flags: HttpOnly=%v Secure=%v", originalCookie.HttpOnly, originalCookie.Secure)
	}

	userEndpoint := app.protected(domain.RoleUser)
	if response := requestWithToken(userEndpoint, originalCookie); response.Code != http.StatusNoContent {
		t.Fatalf("valid token status = %d; body = %s", response.Code, response.Body.String())
	}

	app.users.incrementAuthVersion()
	assertRevocationError(
		t,
		requestWithToken(userEndpoint, originalCookie),
		http.StatusUnauthorized,
		domain.CodeUnauthorized,
	)

	relogin, replacementCookie := app.login(t, revocationTestPassword)
	if relogin.Code != http.StatusOK || replacementCookie == nil {
		t.Fatalf("re-login response = %d %s", relogin.Code, relogin.Body.String())
	}
	if replacementCookie.Value == originalCookie.Value {
		t.Fatal("re-login reused the revoked JWT")
	}
	if response := requestWithToken(userEndpoint, replacementCookie); response.Code != http.StatusNoContent {
		t.Fatalf("replacement token status = %d; body = %s", response.Code, response.Body.String())
	}
}

func TestDisabledAndDeletedAccountsCannotAuthenticateOrUseJWT(t *testing.T) {
	t.Run("disabled account", func(t *testing.T) {
		app := newRevocationTestApplication(t)
		login, cookie := app.login(t, revocationTestPassword)
		if login.Code != http.StatusOK || cookie == nil {
			t.Fatalf("initial login response = %d %s", login.Code, login.Body.String())
		}

		app.users.setDisabled(true)
		assertRevocationError(
			t,
			requestWithToken(app.protected(domain.RoleUser), cookie),
			http.StatusUnauthorized,
			domain.CodeAccountDisabled,
		)

		disabledLogin, disabledCookie := app.login(t, revocationTestPassword)
		if disabledCookie != nil {
			t.Fatal("disabled account received an access token")
		}
		assertRevocationError(
			t,
			disabledLogin,
			http.StatusUnauthorized,
			domain.CodeAccountDisabled,
		)
	})

	t.Run("deleted account", func(t *testing.T) {
		app := newRevocationTestApplication(t)
		login, cookie := app.login(t, revocationTestPassword)
		if login.Code != http.StatusOK || cookie == nil {
			t.Fatalf("initial login response = %d %s", login.Code, login.Body.String())
		}

		app.users.delete()
		assertRevocationError(
			t,
			requestWithToken(app.protected(domain.RoleUser), cookie),
			http.StatusUnauthorized,
			domain.CodeUnauthorized,
		)

		deletedLogin, deletedCookie := app.login(t, revocationTestPassword)
		if deletedCookie != nil {
			t.Fatal("deleted account received an access token")
		}
		assertRevocationError(
			t,
			deletedLogin,
			http.StatusUnauthorized,
			domain.CodeInvalidCredentials,
		)
	})
}

func TestPasswordChangeInvalidatesExistingJWT(t *testing.T) {
	app := newRevocationTestApplication(t)
	login, oldCookie := app.login(t, revocationTestPassword)
	if login.Code != http.StatusOK || oldCookie == nil {
		t.Fatalf("initial login response = %d %s", login.Code, login.Body.String())
	}

	app.users.changePassword(t, revocationTestNewPassword)
	assertRevocationError(
		t,
		requestWithToken(app.protected(domain.RoleUser), oldCookie),
		http.StatusUnauthorized,
		domain.CodeUnauthorized,
	)

	oldPasswordLogin, oldPasswordCookie := app.login(t, revocationTestPassword)
	if oldPasswordCookie != nil {
		t.Fatal("old password received an access token")
	}
	assertRevocationError(
		t,
		oldPasswordLogin,
		http.StatusUnauthorized,
		domain.CodeInvalidCredentials,
	)

	newPasswordLogin, newCookie := app.login(t, revocationTestNewPassword)
	if newPasswordLogin.Code != http.StatusOK || newCookie == nil {
		t.Fatalf("new-password login response = %d %s", newPasswordLogin.Code, newPasswordLogin.Body.String())
	}
	if response := requestWithToken(app.protected(domain.RoleUser), newCookie); response.Code != http.StatusNoContent {
		t.Fatalf("new-password token status = %d; body = %s", response.Code, response.Body.String())
	}
}

func TestRoleChangeRejectsOldJWTAndUsesCurrentRoleAfterRelogin(t *testing.T) {
	app := newRevocationTestApplication(t)
	login, userCookie := app.login(t, revocationTestPassword)
	if login.Code != http.StatusOK || userCookie == nil {
		t.Fatalf("initial login response = %d %s", login.Code, login.Body.String())
	}

	app.users.changeRole(domain.RoleAdmin)
	assertRevocationError(
		t,
		requestWithToken(app.protected(domain.RoleUser), userCookie),
		http.StatusUnauthorized,
		domain.CodeUnauthorized,
	)

	adminLogin, adminCookie := app.login(t, revocationTestPassword)
	if adminLogin.Code != http.StatusOK || adminCookie == nil {
		t.Fatalf("admin re-login response = %d %s", adminLogin.Code, adminLogin.Body.String())
	}
	var loginResponse LoginResponse
	if err := json.Unmarshal(adminLogin.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("decode admin login response: %v", err)
	}
	if loginResponse.Role != domain.RoleAdmin {
		t.Fatalf("login role = %q, want %q", loginResponse.Role, domain.RoleAdmin)
	}
	if response := requestWithToken(app.protected(domain.RoleAdmin), adminCookie); response.Code != http.StatusNoContent {
		t.Fatalf("admin endpoint status = %d; body = %s", response.Code, response.Body.String())
	}
	assertRevocationError(
		t,
		requestWithToken(app.protected(domain.RoleUser), adminCookie),
		http.StatusForbidden,
		domain.CodeForbidden,
	)
}
