package domain

import (
	"time"

	"github.com/google/uuid"
)

type JuryTeam struct {
	TeamID        uuid.UUID
	TeamName      string
	SolutionURL   string
	EvaluatedByMe bool
	MembersCount  int
	SubmissionAt  time.Time
}

type AdminStats struct {
	TotalUsers         int
	TotalTeams         int
	SubmittedSolutions int
	TotalJuries        int
	EvaluationsClosed  bool
}

type ExportSummaryRow struct {
	TeamID           uuid.UUID
	TeamName         string
	CaptainName      string
	CaptainTelegram  string
	SolutionURL      string
	TotalMembers     int
	Members          string
	TotalScore       int
	EvaluatedByCount int
}

type ExportDetailRow struct {
	TeamName    string
	JuryName    string
	CriterionID int
	Score       int
}
