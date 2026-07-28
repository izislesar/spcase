package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	// MinTeamMembers is the minimum number of participants needed to submit a solution.
	MinTeamMembers = 2
	// MaxTeamMembers is the maximum number of participants allowed in a team.
	MaxTeamMembers = 4
)

// TeamStatusBadge describes the team's readiness as shown to participants.
type TeamStatusBadge string

const (
	TeamStatusSearching TeamStatusBadge = "SEARCHING"
	TeamStatusReady     TeamStatusBadge = "READY"
	TeamStatusSubmitted TeamStatusBadge = "SUBMITTED"
)

// IsValid reports whether the status badge is supported by the application.
func (s TeamStatusBadge) IsValid() bool {
	switch s {
	case TeamStatusSearching, TeamStatusReady, TeamStatusSubmitted:
		return true
	default:
		return false
	}
}

// Team is a participant group managed by its captain.
type Team struct {
	ID         uuid.UUID
	Name       string
	InviteCode string
	CaptainID  uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TeamMembership is the persisted association between one USER and one team.
type TeamMembership struct {
	TeamID   uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}

// TeamMember is the public participant representation returned with a team.
type TeamMember struct {
	ID        uuid.UUID
	FullName  string
	Telegram  string
	IsCaptain bool
}

// IsCaptain reports whether userID owns the team.
func (t Team) IsCaptain(userID uuid.UUID) bool {
	return t.CaptainID == userID
}

// HasCapacity reports whether another participant may join the team.
func HasCapacity(memberCount int) bool {
	return memberCount < MaxTeamMembers
}

// SubmissionAllowed reports whether the team has enough members to submit a solution.
func SubmissionAllowed(memberCount int) bool {
	return memberCount >= MinTeamMembers && memberCount <= MaxTeamMembers
}

// StatusBadge derives the team's display status from its member count and
// whether a solution has been submitted.
func StatusBadge(memberCount int, hasSubmission bool) TeamStatusBadge {
	if hasSubmission {
		return TeamStatusSubmitted
	}
	if SubmissionAllowed(memberCount) {
		return TeamStatusReady
	}
	return TeamStatusSearching
}

// TeamState keeps independently derived readiness and mutation lock state.
type TeamState struct {
	Status          TeamStatusBadge
	MutationsLocked bool
}
