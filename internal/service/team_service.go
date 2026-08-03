package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

const (
	InviteCodeLength        = 8
	HardLockBeforeDeadline  = time.Hour
	maximumTeamNameLength   = 100
	inviteGenerationRetries = 8
)

const inviteCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type TeamRepository interface {
	Create(context.Context, domain.Team) (domain.Team, error)
	GetByID(context.Context, uuid.UUID) (domain.Team, error)
	GetByUserID(context.Context, uuid.UUID) (domain.Team, domain.TeamMembership, error)
	ListMembers(context.Context, uuid.UUID) ([]domain.TeamMember, error)
	Join(context.Context, uuid.UUID, string) (domain.Team, error)
	Leave(context.Context, uuid.UUID, time.Time) error
	Kick(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	TransferOwnership(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	Disband(context.Context, uuid.UUID, time.Time) error
}

type TeamSubmissionRepository interface {
	GetByTeamID(context.Context, uuid.UUID) (domain.Submission, error)
}

type TeamDetails struct {
	Team            domain.Team
	Members         []domain.TeamMember
	Status          domain.TeamStatusBadge
	MutationsLocked bool
	Submission      *domain.Submission
}

type TeamService struct {
	teams              TeamRepository
	submissions        TeamSubmissionRepository
	submissionDeadline time.Time
	now                func() time.Time
	generateInviteCode func() (string, error)
}

func NewTeamService(
	teams TeamRepository,
	submissions TeamSubmissionRepository,
	submissionDeadline time.Time,
) (*TeamService, error) {
	if teams == nil || submissions == nil {
		return nil, errors.New("team repositories cannot be nil")
	}
	if submissionDeadline.IsZero() {
		return nil, errors.New("submission deadline cannot be zero")
	}
	return &TeamService{
		teams:              teams,
		submissions:        submissions,
		submissionDeadline: submissionDeadline.UTC(),
		now:                time.Now,
		generateInviteCode: GenerateInviteCode,
	}, nil
}

func (s *TeamService) Create(ctx context.Context, captainID uuid.UUID, name string) (domain.Team, error) {
	if captainID == uuid.Nil {
		return domain.Team{}, domain.ErrUserNotFound
	}
	name = strings.TrimSpace(name)
	if strings.ContainsRune(name, '\x00') {
		return domain.Team{}, &ValidationError{Field: "name", Reason: "must not contain NUL characters"}
	}
	if name == "" {
		return domain.Team{}, &ValidationError{Field: "name", Reason: "must not be empty"}
	}
	if utf8.RuneCountInString(name) > maximumTeamNameLength {
		return domain.Team{}, &ValidationError{Field: "name", Reason: "must not exceed 100 characters"}
	}
	for attempt := 0; attempt < inviteGenerationRetries; attempt++ {
		inviteCode, err := s.generateInviteCode()
		if err != nil {
			return domain.Team{}, fmt.Errorf("generate invite code: %w", err)
		}
		team, err := s.teams.Create(ctx, domain.Team{
			Name: name, InviteCode: inviteCode, CaptainID: captainID,
		})
		if err == nil {
			return team, nil
		}
		if !errors.Is(err, domain.ErrInviteCodeCollision) {
			return domain.Team{}, err
		}
	}
	return domain.Team{}, domain.ErrInviteCodeCollision
}

func (s *TeamService) Join(ctx context.Context, userID uuid.UUID, inviteCode string) (domain.Team, error) {
	if userID == uuid.Nil {
		return domain.Team{}, domain.ErrUserNotFound
	}
	inviteCode = strings.ToUpper(strings.TrimSpace(inviteCode))
	if !validInviteCode(inviteCode) {
		return domain.Team{}, domain.ErrInvalidInviteCode
	}
	return s.teams.Join(ctx, userID, inviteCode)
}

func (s *TeamService) MyTeam(ctx context.Context, userID uuid.UUID) (TeamDetails, error) {
	if userID == uuid.Nil {
		return TeamDetails{}, domain.ErrUserNotFound
	}
	team, _, err := s.teams.GetByUserID(ctx, userID)
	if err != nil {
		return TeamDetails{}, err
	}
	members, err := s.teams.ListMembers(ctx, team.ID)
	if err != nil {
		return TeamDetails{}, err
	}
	submission, err := s.submissions.GetByTeamID(ctx, team.ID)
	var submissionPointer *domain.Submission
	switch {
	case err == nil:
		submissionPointer = &submission
	case errors.Is(err, domain.ErrSubmissionNotFound):
	default:
		return TeamDetails{}, err
	}
	return TeamDetails{
		Team:            team,
		Members:         members,
		Status:          domain.StatusBadge(len(members), submissionPointer != nil),
		MutationsLocked: s.MutationsLocked(),
		Submission:      submissionPointer,
	}, nil
}

func (s *TeamService) Leave(ctx context.Context, userID uuid.UUID) error {
	if err := s.ensureMutationsOpen(); err != nil {
		return err
	}
	return s.teams.Leave(ctx, userID, s.lockAt())
}

func (s *TeamService) Kick(ctx context.Context, captainID, memberID uuid.UUID) error {
	if err := s.ensureMutationsOpen(); err != nil {
		return err
	}
	if memberID == uuid.Nil {
		return domain.ErrTeamMemberNotFound
	}
	return s.teams.Kick(ctx, captainID, memberID, s.lockAt())
}

func (s *TeamService) TransferOwnership(
	ctx context.Context,
	captainID, newCaptainID uuid.UUID,
) error {
	if err := s.ensureMutationsOpen(); err != nil {
		return err
	}
	if newCaptainID == uuid.Nil {
		return domain.ErrTeamMemberNotFound
	}
	return s.teams.TransferOwnership(ctx, captainID, newCaptainID, s.lockAt())
}

func (s *TeamService) Disband(ctx context.Context, captainID uuid.UUID) error {
	if err := s.ensureMutationsOpen(); err != nil {
		return err
	}
	return s.teams.Disband(ctx, captainID, s.lockAt())
}

func (s *TeamService) MutationsLocked() bool {
	return !s.now().UTC().Before(s.lockAt())
}

func (s *TeamService) ensureMutationsOpen() error {
	if s.MutationsLocked() {
		return domain.ErrMutationsLocked
	}
	return nil
}

func (s *TeamService) lockAt() time.Time {
	return s.submissionDeadline.Add(-HardLockBeforeDeadline)
}

func GenerateInviteCode() (string, error) {
	code := make([]byte, InviteCodeLength)
	upperBound := big.NewInt(int64(len(inviteCodeAlphabet)))
	for index := range code {
		value, err := rand.Int(rand.Reader, upperBound)
		if err != nil {
			return "", err
		}
		code[index] = inviteCodeAlphabet[value.Int64()]
	}
	return string(code), nil
}

func validInviteCode(code string) bool {
	if len(code) != InviteCodeLength {
		return false
	}
	for _, character := range code {
		if !strings.ContainsRune(inviteCodeAlphabet, character) {
			return false
		}
	}
	return true
}
