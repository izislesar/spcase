// Package service contains application business logic.
package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"spcase.ru/backend/internal/domain"
)

const (
	JWTExpiration          = 24 * time.Hour
	MinimumPasswordLength  = 8
	maximumPasswordBytes   = 72
	maximumEmailLength     = 255
	maximumProfileLength   = 255
	maximumTelegramLength  = 100
	accessTokenSigningAlgo = "HS256"
)

var dummyPasswordHash = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")

type AuthUserRepository interface {
	Create(context.Context, domain.User) (domain.User, error)
	GetByEmail(context.Context, string) (domain.User, error)
	GetAccountProjection(context.Context, uuid.UUID) (domain.AccountProjection, error)
}

type RegisterInput struct {
	FullName   string
	University string
	Email      string
	Telegram   string
	Password   string
}

type JuryRegisterInput struct {
	SecretKey string
	FullName  string
	Email     string
	Password  string
}

type AuthResult struct {
	User      domain.User
	Token     string
	ExpiresAt time.Time
}

type AccessTokenClaims struct {
	Role        domain.Role `json:"role"`
	AuthVersion int         `json:"ver"`
	jwt.RegisteredClaims
}

func (c AccessTokenClaims) UserID() (uuid.UUID, error) {
	id, err := uuid.Parse(c.Subject)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domain.ErrUnauthorized
	}
	return id, nil
}

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

type AuthService struct {
	users                AuthUserRepository
	jwtSecret            []byte
	juryRegistrationKey  []byte
	registrationDeadline time.Time
	now                  func() time.Time
}

func NewAuthService(
	users AuthUserRepository,
	jwtSecret string,
	juryRegistrationKey string,
	registrationDeadline time.Time,
) (*AuthService, error) {
	if users == nil {
		return nil, errors.New("auth user repository cannot be nil")
	}
	if strings.TrimSpace(jwtSecret) == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}
	if juryRegistrationKey == "" {
		return nil, errors.New("jury registration key cannot be empty")
	}
	if registrationDeadline.IsZero() {
		return nil, errors.New("registration deadline cannot be zero")
	}
	return &AuthService{
		users:                users,
		jwtSecret:            []byte(jwtSecret),
		juryRegistrationKey:  []byte(juryRegistrationKey),
		registrationDeadline: registrationDeadline.UTC(),
		now:                  time.Now,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	if !s.now().UTC().Before(s.registrationDeadline) {
		return AuthResult{}, domain.ErrRegistrationClosed
	}
	normalized, err := normalizeRegisterInput(input)
	if err != nil {
		return AuthResult{}, err
	}
	university, telegram := normalized.University, normalized.Telegram
	user, err := s.createAccount(ctx, domain.User{
		FullName:   normalized.FullName,
		University: &university,
		Email:      normalized.Email,
		Telegram:   &telegram,
		Role:       domain.RoleUser,
	}, normalized.Password)
	if err != nil {
		return AuthResult{}, err
	}
	return s.authenticate(user)
}

func (s *AuthService) RegisterJury(ctx context.Context, input JuryRegisterInput) (AuthResult, error) {
	providedKeyHash := sha256.Sum256([]byte(input.SecretKey))
	expectedKeyHash := sha256.Sum256(s.juryRegistrationKey)
	if subtle.ConstantTimeCompare(providedKeyHash[:], expectedKeyHash[:]) != 1 {
		return AuthResult{}, domain.ErrInvalidSecretKey
	}
	fullName := strings.TrimSpace(input.FullName)
	if fullName == "" || utf8.RuneCountInString(fullName) > maximumProfileLength {
		return AuthResult{}, &ValidationError{Field: "full_name", Reason: "must contain 1 to 255 characters"}
	}
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return AuthResult{}, err
	}
	if err := validatePassword(input.Password); err != nil {
		return AuthResult{}, err
	}
	user, err := s.createAccount(ctx, domain.User{
		FullName: fullName,
		Email:    email,
		Role:     domain.RoleJury,
	}, input.Password)
	if err != nil {
		return AuthResult{}, err
	}
	return s.authenticate(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthResult, error) {
	return s.login(ctx, email, password, domain.RoleUser, domain.RoleAdmin)
}

func (s *AuthService) LoginJury(ctx context.Context, email, password string) (AuthResult, error) {
	return s.login(ctx, email, password, domain.RoleJury)
}

func (s *AuthService) login(
	ctx context.Context,
	email, password string,
	allowedRoles ...domain.Role,
) (AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil || password == "" || len([]byte(password)) > maximumPasswordBytes {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	user, err := s.users.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
			return AuthResult{}, domain.ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	roleAllowed := false
	for _, role := range allowedRoles {
		roleAllowed = roleAllowed || user.Role == role
	}
	if !roleAllowed {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	if user.IsDisabled() {
		return AuthResult{}, domain.ErrAccountDisabled
	}
	return s.authenticate(user)
}

func (s *AuthService) GenerateToken(user domain.User) (string, time.Time, error) {
	if user.ID == uuid.Nil || !user.Role.IsValid() || user.AuthVersion < 1 || user.IsDisabled() {
		return "", time.Time{}, domain.ErrUnauthorized
	}
	issuedAt := s.now().UTC()
	expiresAt := issuedAt.Add(JWTExpiration)
	claims := AccessTokenClaims{
		Role:        user.Role,
		AuthVersion: user.AuthVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (s *AuthService) ValidateToken(rawToken string) (AccessTokenClaims, error) {
	if strings.TrimSpace(rawToken) == "" {
		return AccessTokenClaims{}, domain.ErrUnauthorized
	}
	claims := AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(
		rawToken,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != accessTokenSigningAlgo {
				return nil, domain.ErrUnauthorized
			}
			return s.jwtSecret, nil
		},
		jwt.WithValidMethods([]string{accessTokenSigningAlgo}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(func() time.Time { return s.now().UTC() }),
	)
	if err != nil || !token.Valid || !claims.Role.IsValid() || claims.AuthVersion < 1 {
		return AccessTokenClaims{}, domain.ErrUnauthorized
	}
	if _, err := claims.UserID(); err != nil {
		return AccessTokenClaims{}, err
	}
	return claims, nil
}

func (s *AuthService) VerifyAccount(
	ctx context.Context,
	claims AccessTokenClaims,
) (domain.AccountProjection, error) {
	userID, err := claims.UserID()
	if err != nil {
		return domain.AccountProjection{}, err
	}
	projection, err := s.users.GetAccountProjection(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.AccountProjection{}, domain.ErrUnauthorized
		}
		return domain.AccountProjection{}, err
	}
	if projection.DisabledAt != nil {
		return domain.AccountProjection{}, domain.ErrAccountDisabled
	}
	if projection.AuthVersion != claims.AuthVersion || projection.Role != claims.Role {
		return domain.AccountProjection{}, domain.ErrUnauthorized
	}
	return projection, nil
}

func (s *AuthService) createAccount(
	ctx context.Context,
	user domain.User,
	password string,
) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user.PasswordHash = string(hash)
	return s.users.Create(ctx, user)
}

func (s *AuthService) authenticate(user domain.User) (AuthResult, error) {
	token, expiresAt, err := s.GenerateToken(user)
	if err != nil {
		return AuthResult{}, err
	}
	user.PasswordHash = ""
	return AuthResult{User: user, Token: token, ExpiresAt: expiresAt}, nil
}

func normalizeRegisterInput(input RegisterInput) (RegisterInput, error) {
	input.FullName = strings.TrimSpace(input.FullName)
	input.University = strings.TrimSpace(input.University)
	input.Telegram = strings.TrimSpace(input.Telegram)
	for field, value := range map[string]string{
		"full_name": input.FullName, "university": input.University, "telegram": input.Telegram,
	} {
		if value == "" {
			return RegisterInput{}, &ValidationError{Field: field, Reason: "must not be empty"}
		}
	}
	if utf8.RuneCountInString(input.FullName) > maximumProfileLength ||
		utf8.RuneCountInString(input.University) > maximumProfileLength {
		return RegisterInput{}, &ValidationError{Field: "profile", Reason: "must not exceed 255 characters"}
	}
	if utf8.RuneCountInString(input.Telegram) > maximumTelegramLength {
		return RegisterInput{}, &ValidationError{Field: "telegram", Reason: "must not exceed 100 characters"}
	}
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return RegisterInput{}, err
	}
	input.Email = email
	if err := validatePassword(input.Password); err != nil {
		return RegisterInput{}, err
	}
	return input, nil
}

func validatePassword(password string) error {
	length := len([]byte(password))
	if length < MinimumPasswordLength {
		return &ValidationError{Field: "password", Reason: "must contain at least 8 bytes"}
	}
	if length > maximumPasswordBytes {
		return &ValidationError{Field: "password", Reason: "must not exceed 72 bytes"}
	}
	return nil
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maximumEmailLength {
		return "", &ValidationError{Field: "email", Reason: "must be a valid email address"}
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return "", &ValidationError{Field: "email", Reason: "must be a valid email address"}
	}
	return email, nil
}
