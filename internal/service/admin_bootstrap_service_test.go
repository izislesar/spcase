package service_test

import (
	"context"
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

type bootstrapAuthRepository struct {
	mu    sync.Mutex
	admin *domain.User
}

func (r *bootstrapAuthRepository) CreateFirstAdmin(
	_ context.Context,
	admin domain.User,
) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.admin != nil {
		return domain.User{}, domain.ErrAdminAlreadyExists
	}
	admin.ID = uuid.New()
	admin.AuthVersion = 1
	admin.CreatedAt = time.Now().UTC()
	r.admin = &admin
	return admin, nil
}

func (r *bootstrapAuthRepository) Create(
	context.Context,
	domain.User,
) (domain.User, error) {
	return domain.User{}, errors.New("unexpected account creation")
}

func (r *bootstrapAuthRepository) GetByEmail(
	_ context.Context,
	email string,
) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.admin == nil || !strings.EqualFold(r.admin.Email, email) {
		return domain.User{}, domain.ErrUserNotFound
	}
	return *r.admin, nil
}

func (r *bootstrapAuthRepository) GetAccountProjection(
	_ context.Context,
	id uuid.UUID,
) (domain.AccountProjection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.admin == nil || r.admin.ID != id {
		return domain.AccountProjection{}, domain.ErrUserNotFound
	}
	return domain.AccountProjection{
		ID:          r.admin.ID,
		Role:        r.admin.Role,
		AuthVersion: r.admin.AuthVersion,
		DisabledAt:  r.admin.DisabledAt,
	}, nil
}

func TestAdminBootstrapCreatesOneLoginCapableAdministrator(t *testing.T) {
	const (
		email    = "initial-admin@example.com"
		password = "bootstrap-password-42"
	)
	users := &bootstrapAuthRepository{}
	bootstrap, err := service.NewAdminBootstrapService(users)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := bootstrap.Bootstrap(context.Background(), service.AdminBootstrapInput{
		FullName: "Initial Administrator",
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	if admin.Role != domain.RoleAdmin || admin.PasswordHash != "" {
		t.Fatalf("unexpected bootstrap result: %#v", admin)
	}
	if users.admin == nil {
		t.Fatal("administrator was not persisted")
	}
	if users.admin.PasswordHash == password {
		t.Fatal("plaintext password was persisted")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(users.admin.PasswordHash),
		[]byte(password),
	); err != nil {
		t.Fatalf("stored bcrypt hash does not verify: %v", err)
	}

	_, err = bootstrap.Bootstrap(context.Background(), service.AdminBootstrapInput{
		FullName: "Second Administrator",
		Email:    "second-admin@example.com",
		Password: "another-bootstrap-password-42",
	})
	if !errors.Is(err, domain.ErrAdminAlreadyExists) {
		t.Fatalf("second bootstrap error = %v", err)
	}

	auth, err := service.NewAuthService(
		users,
		"unit-test-jwt-secret-that-is-long-enough",
		"unit-test-jury-key",
		time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	login, err := auth.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("login with bootstrapped credentials: %v", err)
	}
	if login.User.Role != domain.RoleAdmin || login.User.PasswordHash != "" {
		t.Fatalf("unexpected login result: %#v", login.User)
	}
	claims, err := auth.ValidateToken(login.Token)
	if err != nil {
		t.Fatalf("validate admin token: %v", err)
	}
	if claims.Role != domain.RoleAdmin {
		t.Fatalf("token role = %q", claims.Role)
	}

	authMiddleware, err := httpmiddleware.NewAuthMiddleware(auth, nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := authMiddleware.Middleware(
		httpmiddleware.RequireRoles(domain.RoleAdmin)(
			http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusNoContent)
			}),
		),
	)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	request.AddCookie(&http.Cookie{
		Name:  httpmiddleware.AccessTokenCookieName,
		Value: login.Token,
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("ADMIN permission status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
