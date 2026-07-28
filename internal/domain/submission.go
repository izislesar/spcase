package domain

import (
	"time"

	"github.com/google/uuid"
)

// Submission is the current solution URL of a team.
type Submission struct {
	ID          uuid.UUID
	TeamID      uuid.UUID
	SolutionURL string
	UpdatedAt   time.Time
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
