package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	// FirstCriterionID is the first criterion available for jury evaluation.
	FirstCriterionID CriterionID = 1
	// LastCriterionID is the last criterion in the fixed six-criterion rubric.
	LastCriterionID CriterionID = 6
	// MinScore is the lowest score a jury member can assign to a criterion.
	MinScore Score = 0
	// MaxScore is the highest score a jury member can assign to a criterion.
	MaxScore Score = 10
)

// CriterionID identifies one criterion in the evaluation rubric.
type CriterionID int

// IsValid reports whether the criterion belongs to the fixed rubric.
func (id CriterionID) IsValid() bool {
	return id >= FirstCriterionID && id <= LastCriterionID
}

// Score is a jury score for one criterion.
type Score int

// IsValid reports whether the score falls within the permitted inclusive range.
func (s Score) IsValid() bool {
	return s >= MinScore && s <= MaxScore
}

// Evaluation is one jury member's score for one team criterion. The
// JuryID-TeamID-CriterionID tuple is unique.
type Evaluation struct {
	ID          uuid.UUID
	JuryID      uuid.UUID
	TeamID      uuid.UUID
	CriterionID CriterionID
	Score       Score
	UpdatedAt   time.Time
}

// IsValid reports whether the evaluation has valid criterion and score values.
func (e Evaluation) IsValid() bool {
	return e.CriterionID.IsValid() && e.Score.IsValid()
}

// EvaluationTotal returns the sum of valid evaluation scores. Invalid scores
// are ignored because they cannot contribute to a persisted aggregate.
func EvaluationTotal(evaluations []Evaluation) Score {
	var total Score
	for _, evaluation := range evaluations {
		if evaluation.IsValid() {
			total += evaluation.Score
		}
	}
	return total
}
