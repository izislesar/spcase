package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role is the global RBAC role of an account.
type Role string

const (
	RoleUser  Role = "USER"
	RoleJury  Role = "JURY"
	RoleAdmin Role = "ADMIN"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleJury, RoleAdmin:
		return true
	default:
		return false
	}
}

// MembershipState is a USER account's relationship to a team. It is not an
// RBAC role and is never stored in JWT claims.
type MembershipState string

const (
	MembershipNoTeam  MembershipState = "NO_TEAM"
	MembershipInTeam  MembershipState = "IN_TEAM"
	MembershipCaptain MembershipState = "CAPTAIN"
)

func (s MembershipState) IsValid() bool {
	switch s {
	case MembershipNoTeam, MembershipInTeam, MembershipCaptain:
		return true
	default:
		return false
	}
}

// User is the single identity model used by participants, jury, and admins.
// University and Telegram are required only for RoleUser.
type User struct {
	ID           uuid.UUID
	FullName     string
	University   *string
	Email        string
	Telegram     *string
	PasswordHash string
	Role         Role
	AuthVersion  int
	DisabledAt   *time.Time
	CreatedAt    time.Time
}

// AccountProjection is the minimal mutable identity data checked on every
// authenticated request.
type AccountProjection struct {
	ID          uuid.UUID
	Role        Role
	AuthVersion int
	DisabledAt  *time.Time
}

func (u User) IsDisabled() bool {
	return u.DisabledAt != nil
}

// MembershipStateFor derives membership state from current database data.
func MembershipStateFor(userID uuid.UUID, membership *TeamMembership, captainID uuid.UUID) MembershipState {
	if membership == nil {
		return MembershipNoTeam
	}
	if userID == captainID {
		return MembershipCaptain
	}
	return MembershipInTeam
}
