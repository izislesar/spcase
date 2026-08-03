package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"spcase.ru/backend/internal/domain"
)

const securityTestJWTSecret = "security-test-jwt-secret-that-is-long-enough"

func TestValidateTokenRejectsJWTAbuse(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	users := &fakeAuthUsers{}
	auth, err := NewAuthService(users, securityTestJWTSecret, "jury-secret", now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	auth.now = func() time.Time { return now }

	validClaims := func() AccessTokenClaims {
		return AccessTokenClaims{
			Role: domain.RoleUser, AuthVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(), IssuedAt: jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			},
		}
	}
	sign := func(t *testing.T, method jwt.SigningMethod, claims AccessTokenClaims, key any) string {
		t.Helper()
		token, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}
		return token
	}

	valid := sign(t, jwt.SigningMethodHS256, validClaims(), []byte(securityTestJWTSecret))
	if _, err := auth.ValidateToken(valid); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}

	tests := []struct {
		name  string
		token func(*testing.T) string
	}{
		{name: "expired", token: func(t *testing.T) string {
			claims := validClaims()
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Second))
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "missing expiration", token: func(t *testing.T) string {
			claims := validClaims()
			claims.ExpiresAt = nil
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "future not before", token: func(t *testing.T) string {
			claims := validClaims()
			claims.NotBefore = jwt.NewNumericDate(now.Add(time.Minute))
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "future issued at", token: func(t *testing.T) string {
			claims := validClaims()
			claims.IssuedAt = jwt.NewNumericDate(now.Add(time.Minute))
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "wrong secret", token: func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS256, validClaims(), []byte("different-secret"))
		}},
		{name: "unsigned", token: func(t *testing.T) string {
			return sign(t, jwt.SigningMethodNone, validClaims(), jwt.UnsafeAllowNoneSignatureType)
		}},
		{name: "algorithm confusion", token: func(t *testing.T) string {
			return sign(t, jwt.SigningMethodHS384, validClaims(), []byte(securityTestJWTSecret))
		}},
		{name: "unknown role", token: func(t *testing.T) string {
			claims := validClaims()
			claims.Role = domain.Role("ROOT")
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "missing subject", token: func(t *testing.T) string {
			claims := validClaims()
			claims.Subject = ""
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "modified user id", token: func(t *testing.T) string {
			claims := validClaims()
			claims.Subject = uuid.NewString()
			signed := sign(t, jwt.SigningMethodHS256, claims, []byte("attacker-secret"))
			return signed
		}},
		{name: "invalid auth version", token: func(t *testing.T) string {
			claims := validClaims()
			claims.AuthVersion = 0
			return sign(t, jwt.SigningMethodHS256, claims, []byte(securityTestJWTSecret))
		}},
		{name: "malformed", token: func(*testing.T) string { return "not-a-jwt" }},
		{name: "truncated", token: func(*testing.T) string {
			return valid[:len(valid)-1]
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := auth.ValidateToken(test.token(t)); !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("ValidateToken() error = %v, want unauthorized", err)
			}
		})
	}
}

func TestLoginDoesNotDiscloseAccountState(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	disabledAt := time.Now().UTC()

	tests := []struct {
		name       string
		repository *fakeAuthUsers
		password   string
	}{
		{name: "nonexistent account", repository: &fakeAuthUsers{}, password: "wrong-password"},
		{name: "wrong password", repository: &fakeAuthUsers{lookup: &domain.User{
			ID: uuid.New(), Email: "user@example.com", PasswordHash: string(passwordHash),
			Role: domain.RoleUser, AuthVersion: 1,
		}}, password: "wrong-password"},
		{name: "disabled account", repository: &fakeAuthUsers{lookup: &domain.User{
			ID: uuid.New(), Email: "user@example.com", PasswordHash: string(passwordHash),
			Role: domain.RoleUser, AuthVersion: 2, DisabledAt: &disabledAt,
		}}, password: "correct-password"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			auth, err := NewAuthService(
				test.repository, securityTestJWTSecret, "jury-secret", time.Now().Add(time.Hour),
			)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := auth.Login(context.Background(), "user@example.com", test.password); !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want invalid credentials", err)
			}
		})
	}
}

func TestRegistrationRejectsEmbeddedNULBeforePersistence(t *testing.T) {
	t.Parallel()

	users := &fakeAuthUsers{}
	auth, err := NewAuthService(users, securityTestJWTSecret, "jury-secret", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, err = auth.Register(context.Background(), RegisterInput{
		FullName: "Invalid\x00Name", University: "University", Email: "user@example.com",
		Telegram: "@user", Password: "password123",
	})
	var validationError *ValidationError
	if !errors.As(err, &validationError) || !strings.Contains(validationError.Reason, "NUL") {
		t.Fatalf("Register() error = %v, want NUL validation error", err)
	}
	if users.created.ID != uuid.Nil {
		t.Fatal("invalid profile reached repository")
	}
}
