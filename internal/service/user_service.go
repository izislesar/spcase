package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type UserProfileRepository interface {
	GetByID(context.Context, uuid.UUID) (domain.User, error)
}

type UserTeamRepository interface {
	GetByUserID(context.Context, uuid.UUID) (domain.Team, domain.TeamMembership, error)
}

type UserProfile struct {
	User            domain.User
	MembershipState domain.MembershipState
	TeamID          *uuid.UUID
}

type UserService struct {
	users UserProfileRepository
	teams UserTeamRepository
}

func NewUserService(users UserProfileRepository, teams UserTeamRepository) (*UserService, error) {
	if users == nil || teams == nil {
		return nil, errors.New("user service repositories cannot be nil")
	}
	return &UserService{users: users, teams: teams}, nil
}

func (s *UserService) Me(ctx context.Context, userID uuid.UUID) (UserProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return UserProfile{}, err
	}
	user.PasswordHash = ""
	team, membership, err := s.teams.GetByUserID(ctx, userID)
	if errors.Is(err, domain.ErrNoTeam) {
		return UserProfile{User: user, MembershipState: domain.MembershipNoTeam}, nil
	}
	if err != nil {
		return UserProfile{}, err
	}
	return UserProfile{
		User: user, MembershipState: domain.MembershipStateFor(userID, &membership, team.CaptainID),
		TeamID: &team.ID,
	}, nil
}
