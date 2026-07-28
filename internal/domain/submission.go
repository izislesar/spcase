package domain

import (
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaximumSolutionURLLength is the largest accepted solution URL in bytes.
const MaximumSolutionURLLength = 2048

// Submission is the current solution URL of a team.
type Submission struct {
	ID          uuid.UUID
	TeamID      uuid.UUID
	SolutionURL string
	UpdatedAt   time.Time
}

// NormalizeSolutionURL validates and returns a canonical HTTP(S) solution URL.
func NormalizeSolutionURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > MaximumSolutionURLLength {
		return "", ErrInvalidURLFormat
	}

	parsed, err := url.ParseRequestURI(raw)
	if err != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Hostname() == "" {
		return "", ErrInvalidURLFormat
	}
	return parsed.String(), nil
}

// EvaluationState is the singleton switch controlled by administrators.
type EvaluationState struct {
	IsClosed  bool
	ClosedAt  *time.Time
	ClosedBy  *uuid.UUID
	UpdatedAt time.Time
}

type EvaluationStateAction string

const (
	EvaluationOpened EvaluationStateAction = "OPEN"
	EvaluationClosed EvaluationStateAction = "CLOSE"
)

// EvaluationStateEvent is an append-only audit record.
type EvaluationStateEvent struct {
	ID        uuid.UUID
	Action    EvaluationStateAction
	AdminID   uuid.UUID
	CreatedAt time.Time
}
