package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"spcase.ru/backend/internal/domain"
)

type AdminBootstrapRepository interface {
	CreateFirstAdmin(context.Context, domain.User) (domain.User, error)
}

type AdminBootstrapInput struct {
	FullName string
	Email    string
	Password string
}

type AdminBootstrapService struct {
	users AdminBootstrapRepository
}

func NewAdminBootstrapService(users AdminBootstrapRepository) (*AdminBootstrapService, error) {
	if users == nil {
		return nil, errors.New("admin bootstrap repository cannot be nil")
	}
	return &AdminBootstrapService{users: users}, nil
}

func (s *AdminBootstrapService) Bootstrap(
	ctx context.Context,
	input AdminBootstrapInput,
) (domain.User, error) {
	fullName := strings.TrimSpace(input.FullName)
	if strings.ContainsRune(fullName, '\x00') {
		return domain.User{}, &ValidationError{
			Field: "full_name", Reason: "must not contain NUL characters",
		}
	}
	if fullName == "" || utf8.RuneCountInString(fullName) > maximumProfileLength {
		return domain.User{}, &ValidationError{
			Field: "full_name", Reason: "must contain 1 to 255 characters",
		}
	}
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return domain.User{}, err
	}
	if err := validatePassword(input.Password); err != nil {
		return domain.User{}, err
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return domain.User{}, err
	}

	admin, err := s.users.CreateFirstAdmin(ctx, domain.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.RoleAdmin,
	})
	if err != nil {
		return domain.User{}, err
	}
	admin.PasswordHash = ""
	return admin, nil
}
