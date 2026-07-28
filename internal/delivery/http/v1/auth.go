package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/service"
)

// AuthenticationService contains authentication use cases used by AuthHandler.
type AuthenticationService interface {
	Register(context.Context, service.RegisterInput) (service.AuthResult, error)
	Login(context.Context, string, string) (service.AuthResult, error)
	RegisterJury(context.Context, service.JuryRegisterInput) (service.AuthResult, error)
	LoginJury(context.Context, string, string) (service.AuthResult, error)
}

// RegisterJury creates a jury account using the configured registration key.
func (h *AuthHandler) RegisterJury(writer http.ResponseWriter, request *http.Request) {
	var input JuryRegisterRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	result, err := h.auth.RegisterJury(request.Context(), service.JuryRegisterInput{
		SecretKey: input.SecretKey,
		FullName:  input.FullName,
		Email:     input.Email,
		Password:  input.Password,
	})
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	h.setAccessTokenCookie(writer, result.Token, result.ExpiresAt)
	writeJSON(writer, http.StatusCreated, MessageResponse{Message: "Jury registered successfully"})
}

// LoginJury authenticates only an account with the JURY role.
func (h *AuthHandler) LoginJury(writer http.ResponseWriter, request *http.Request) {
	var input JuryLoginRequest
	if !decodeJSON(writer, request, &input) {
		return
	}
	result, err := h.auth.LoginJury(request.Context(), input.Email, input.Password)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}
	h.setAccessTokenCookie(writer, result.Token, result.ExpiresAt)
	writeJSON(writer, http.StatusOK, LoginResponse{Status: "success", Role: result.User.Role})
}

var _ AuthenticationService = (*service.AuthService)(nil)

// AuthHandler serves participant registration and session endpoints.
type AuthHandler struct {
	auth      AuthenticationService
	appDomain string
	logger    *slog.Logger
}

// NewAuthHandler creates an authentication HTTP handler.
func NewAuthHandler(
	auth AuthenticationService,
	appDomain string,
	logger *slog.Logger,
) (*AuthHandler, error) {
	if auth == nil {
		return nil, errors.New("authentication service cannot be nil")
	}
	return &AuthHandler{
		auth:      auth,
		appDomain: normalizedDomain(appDomain),
		logger:    normalizedLogger(logger),
	}, nil
}

// Register creates a participant and starts an authenticated session.
func (h *AuthHandler) Register(writer http.ResponseWriter, request *http.Request) {
	var input RegisterRequest
	if !decodeJSON(writer, request, &input) {
		return
	}

	result, err := h.auth.Register(request.Context(), service.RegisterInput{
		FullName:   input.FullName,
		University: input.University,
		Email:      input.Email,
		Telegram:   input.Telegram,
		Password:   input.Password,
	})
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	h.setAccessTokenCookie(writer, result.Token, result.ExpiresAt)
	writeJSON(writer, http.StatusCreated, RegisterResponse{
		ID:    result.User.ID,
		Email: result.User.Email,
		Role:  result.User.Role,
	})
}

// Login verifies credentials and starts a new authenticated session.
func (h *AuthHandler) Login(writer http.ResponseWriter, request *http.Request) {
	var input LoginRequest
	if !decodeJSON(writer, request, &input) {
		return
	}

	result, err := h.auth.Login(request.Context(), input.Email, input.Password)
	if err != nil {
		handleError(writer, request, h.logger, err)
		return
	}

	h.setAccessTokenCookie(writer, result.Token, result.ExpiresAt)
	writeJSON(writer, http.StatusOK, LoginResponse{
		Status: "success",
		Role:   result.User.Role,
	})
}

// Logout invalidates the browser's access_token cookie.
func (h *AuthHandler) Logout(writer http.ResponseWriter, request *http.Request) {
	if _, ok := requireAuthenticated(writer, request); !ok {
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     middleware.AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.appDomain,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(writer, http.StatusOK, LogoutResponse{Status: "logged_out"})
}

func (h *AuthHandler) setAccessTokenCookie(writer http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(writer, &http.Cookie{
		Name:     middleware.AccessTokenCookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.appDomain,
		MaxAge:   int(service.JWTExpiration / time.Second),
		Expires:  expiresAt.UTC(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
