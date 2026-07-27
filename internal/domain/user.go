package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the permissions assigned to an authenticated principal.
type Role string

const (
	RoleUser  Role = "USER"
	RoleJury  Role = "JURY"
	RoleAdmin Role = "ADMIN"
)

// IsValid reports whether the role is supported by the application RBAC model.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleJury, RoleAdmin:
		return true
	default:
		return false
	}
}

// TeamStatus describes a participant's relationship to a team.
type TeamStatus string

const (
	TeamStatusNoTeam  TeamStatus = "NO_TEAM"
	TeamStatusInTeam  TeamStatus = "IN_TEAM"
	TeamStatusCaptain TeamStatus = "CAPTAIN"
)

// IsValid reports whether the team status is part of the user state machine.
func (s TeamStatus) IsValid() bool {
	switch s {
	case TeamStatusNoTeam, TeamStatusInTeam, TeamStatusCaptain:
		return true
	default:
		return false
	}
}

// User is a participant or administrator account. TeamID is nil when the user
// is not a member of a team.
type User struct {
	ID           uuid.UUID
	FullName     string
	University   string
	Email        string
	Telegram     string
	PasswordHash string
	Role         Role
	TeamID       *uuid.UUID
	CreatedAt    time.Time
}

// TeamStatus returns the user's current state in the team state machine.
// captainID may be nil when no team has been found.
func (u User) TeamStatus(captainID *uuid.UUID) TeamStatus {
	if u.TeamID == nil {
		return TeamStatusNoTeam
	}
	if captainID != nil && u.ID == *captainID {
		return TeamStatusCaptain
	}
	return TeamStatusInTeam
}

// Jury is an isolated expert account. It deliberately has no team membership
// or participant profile fields.
type Jury struct {
	ID           uuid.UUID
	FullName     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
