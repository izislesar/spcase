// Package v1 contains HTTP request and response contracts for API version 1.
package v1

import (
	"time"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

// ErrorResponse is the unified error response returned by the API.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains a stable error code and a human-readable description.
type ErrorBody struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

// MessageResponse acknowledges a successful command.
type MessageResponse struct {
	Message string `json:"message"`
}

// HealthResponse reports service availability.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ChampionshipInfoResponse provides public championship deadlines.
type ChampionshipInfoResponse struct {
	RegistrationDeadline time.Time `json:"registration_deadline"`
	SubmissionDeadline   time.Time `json:"submission_deadline"`
	IsRegistrationOpen   bool      `json:"is_registration_open"`
	IsSubmissionOpen     bool      `json:"is_submission_open"`
}

// ScheduleEventResponse is one public schedule entry.
type ScheduleEventResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	Description string    `json:"description"`
}

// ScheduleResponse contains all public schedule entries.
type ScheduleResponse struct {
	Events []ScheduleEventResponse `json:"events"`
}

// FAQItemResponse is one FAQ entry.
type FAQItemResponse struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// FAQResponse contains all FAQ entries.
type FAQResponse struct {
	FAQ []FAQItemResponse `json:"faq"`
}

// NoTeamResponse directs a participant to the teammate search chat.
type NoTeamResponse struct {
	Message         string `json:"message"`
	TelegramChatURL string `json:"telegram_chat_url"`
}

// RegisterRequest contains participant registration credentials and profile data.
type RegisterRequest struct {
	FullName   string `json:"full_name"`
	University string `json:"university"`
	Email      string `json:"email"`
	Telegram   string `json:"telegram"`
	Password   string `json:"password"`
}

// RegisterResponse is returned after a participant account is created.
type RegisterResponse struct {
	ID    uuid.UUID   `json:"id"`
	Email string      `json:"email"`
	Role  domain.Role `json:"role"`
}

// LoginRequest contains account credentials.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse acknowledges successful authentication.
type LoginResponse struct {
	Status string      `json:"status"`
	Role   domain.Role `json:"role"`
}

// LogoutResponse acknowledges session termination.
type LogoutResponse struct {
	Status string `json:"status"`
}

// UserMeResponse contains the authenticated participant profile.
type UserMeResponse struct {
	ID         uuid.UUID              `json:"id"`
	FullName   string                 `json:"full_name"`
	University string                 `json:"university"`
	Email      string                 `json:"email"`
	Telegram   string                 `json:"telegram"`
	Role       domain.Role            `json:"role"`
	TeamStatus domain.MembershipState `json:"team_status"`
	TeamID     *uuid.UUID             `json:"team_id"`
}

// CreateTeamRequest contains the desired team name.
type CreateTeamRequest struct {
	Name string `json:"name"`
}

// CreateTeamResponse contains the newly created team identity.
type CreateTeamResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CaptainID  uuid.UUID `json:"captain_id"`
}

// JoinTeamRequest contains a team's invite code.
type JoinTeamRequest struct {
	InviteCode string `json:"invite_code"`
}

// JoinTeamResponse acknowledges a successful team join.
type JoinTeamResponse struct {
	Message  string    `json:"message"`
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
}

// TeamMemberResponse is one member in a team's public roster.
type TeamMemberResponse struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Telegram  string    `json:"telegram"`
	IsCaptain bool      `json:"is_captain"`
}

// TeamSubmissionResponse is a submitted solution attached to a team.
type TeamSubmissionResponse struct {
	SolutionURL string    `json:"solution_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MyTeamResponse contains a team, its roster, and its optional submission.
type MyTeamResponse struct {
	ID              uuid.UUID               `json:"id"`
	Name            string                  `json:"name"`
	InviteCode      string                  `json:"invite_code"`
	CaptainID       uuid.UUID               `json:"captain_id"`
	StatusBadge     domain.TeamStatusBadge  `json:"status_badge"`
	MutationsLocked bool                    `json:"mutations_locked"`
	Members         []TeamMemberResponse    `json:"members"`
	Submission      *TeamSubmissionResponse `json:"submission"`
}

// KickTeamMemberRequest identifies the member a captain wants to remove.
type KickTeamMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

// TransferOwnershipRequest identifies the member who becomes captain.
type TransferOwnershipRequest struct {
	NewCaptainID uuid.UUID `json:"new_captain_id"`
}

// SubmitSolutionRequest contains the URL to save for a team submission.
type SubmitSolutionRequest struct {
	SolutionURL string `json:"solution_url"`
}

// SubmitSolutionResponse contains the persisted solution state.
type SubmitSolutionResponse struct {
	Status      string    `json:"status"`
	SolutionURL string    `json:"solution_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JuryRegisterRequest contains jury registration data and its protected key.
type JuryRegisterRequest struct {
	SecretKey string `json:"secret_key"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// JuryLoginRequest contains jury account credentials.
type JuryLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// JuryTeamResponse is the team view available to a jury member.
type JuryTeamResponse struct {
	TeamID          uuid.UUID `json:"team_id"`
	TeamName        string    `json:"team_name"`
	SolutionURL     string    `json:"solution_url"`
	IsEvaluatedByMe bool      `json:"is_evaluated_by_me"`
	MembersCount    int       `json:"members_count"`
}

// JuryTeamsResponse contains teams visible to the authenticated jury member.
type JuryTeamsResponse struct {
	Teams []JuryTeamResponse `json:"teams"`
}

// EvaluationResponse is one saved score returned to its author.
type EvaluationResponse struct {
	TeamID      uuid.UUID          `json:"team_id"`
	CriterionID domain.CriterionID `json:"criterion_id"`
	Score       domain.Score       `json:"score"`
}

// JuryEvaluationsResponse contains all evaluations created by one jury member.
type JuryEvaluationsResponse struct {
	Evaluations []EvaluationResponse `json:"evaluations"`
}

// CriterionScoreRequest is one criterion score submitted by a jury member.
type CriterionScoreRequest struct {
	CriterionID domain.CriterionID `json:"criterion_id"`
	Score       domain.Score       `json:"score"`
}

// SaveEvaluationsRequest replaces or updates a jury member's scores for one team.
type SaveEvaluationsRequest struct {
	TeamID uuid.UUID               `json:"team_id"`
	Scores []CriterionScoreRequest `json:"scores"`
}

// AdminStatsResponse contains aggregate platform counters.
type AdminStatsResponse struct {
	TotalUsers         int  `json:"total_users"`
	TotalTeams         int  `json:"total_teams"`
	SubmittedSolutions int  `json:"submitted_solutions"`
	TotalJuries        int  `json:"total_juries"`
	EvaluationsClosed  bool `json:"evaluations_closed"`
}
