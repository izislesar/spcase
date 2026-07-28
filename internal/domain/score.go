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
	// CriterionCount is the exact number of rubric criteria.
	CriterionCount = 6
	// MaximumJuryTotal is the maximum total assigned by one jury member.
	MaximumJuryTotal Score = CriterionCount * MaxScore
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

// TeamScoreTotal is a team aggregate with explicit jury coverage.
type TeamScoreTotal struct {
	TeamID           uuid.UUID
	Total            Score
	EvaluatedByCount int
}

// IsValid reports whether the evaluation has valid criterion and score values.
func (e Evaluation) IsValid() bool {
	return e.CriterionID.IsValid() && e.Score.IsValid()
}

// JuryEvaluationTotal validates a complete six-criterion set from one jury
// member for one team and returns its total.
func JuryEvaluationTotal(evaluations []Evaluation) (Score, error) {
	if len(evaluations) != CriterionCount {
		return 0, ErrInvalidEvaluation
	}

	var juryID, teamID uuid.UUID
	criteria := make(map[CriterionID]struct{}, CriterionCount)
	var total Score
	for index, evaluation := range evaluations {
		if evaluation.JuryID == uuid.Nil || evaluation.TeamID == uuid.Nil || !evaluation.IsValid() {
			return 0, ErrInvalidEvaluation
		}
		if index == 0 {
			juryID = evaluation.JuryID
			teamID = evaluation.TeamID
		} else if evaluation.JuryID != juryID || evaluation.TeamID != teamID {
			return 0, ErrInvalidEvaluation
		}
		if _, exists := criteria[evaluation.CriterionID]; exists {
			return 0, ErrInvalidEvaluation
		}
		criteria[evaluation.CriterionID] = struct{}{}
		total += evaluation.Score
	}
	return total, nil
}
