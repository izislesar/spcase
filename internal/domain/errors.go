package domain

// ErrorCode is the stable machine-readable identifier returned for a domain
// rule violation.
type ErrorCode string

const (
	CodeEmailAlreadyExists    ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInvalidCredentials    ErrorCode = "INVALID_CREDENTIALS"
	CodeUnauthorized          ErrorCode = "UNAUTHORIZED"
	CodeForbidden             ErrorCode = "FORBIDDEN"
	CodeUserNotFound          ErrorCode = "USER_NOT_FOUND"
	CodeTeamNotFound          ErrorCode = "TEAM_NOT_FOUND"
	CodeNoTeam                ErrorCode = "NO_TEAM"
	CodeAlreadyInTeam         ErrorCode = "ALREADY_IN_TEAM"
	CodeTeamFull              ErrorCode = "TEAM_FULL"
	CodeInvalidInviteCode     ErrorCode = "INVALID_INVITE_CODE"
	CodeNotTeamCaptain        ErrorCode = "NOT_TEAM_CAPTAIN"
	CodeCaptainCannotLeave    ErrorCode = "CAPTAIN_CANNOT_LEAVE"
	CodeCaptainCannotBeKicked ErrorCode = "CAPTAIN_CANNOT_BE_KICKED"
	CodeTeamMemberNotFound    ErrorCode = "TEAM_MEMBER_NOT_FOUND"
	CodeMutationsLocked       ErrorCode = "MUTATIONS_LOCKED"
	CodeInvalidURLFormat      ErrorCode = "INVALID_URL_FORMAT"
	CodeMinimumTwoMembers     ErrorCode = "MINIMUM_2_MEMBERS_REQUIRED"
	CodeDeadlinePassed        ErrorCode = "DEADLINE_PASSED"
	CodeInvalidSecretKey      ErrorCode = "INVALID_SECRET_KEY"
	CodeInvalidEvaluation     ErrorCode = "INVALID_EVALUATION"
	CodeEvaluationLocked      ErrorCode = "EVALUATIONS_LOCKED"
)

// DomainError represents a business-rule violation that can be safely exposed
// by the HTTP delivery layer.
type DomainError struct {
	Code    ErrorCode
	Message string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	return string(e.Code)
}

// Is allows errors.Is to compare domain errors by their stable code.
func (e *DomainError) Is(target error) bool {
	targetDomainError, ok := target.(*DomainError)
	return ok && e.Code == targetDomainError.Code
}

var (
	ErrEmailAlreadyExists    = &DomainError{Code: CodeEmailAlreadyExists, Message: "Email is already registered"}
	ErrInvalidCredentials    = &DomainError{Code: CodeInvalidCredentials, Message: "Invalid email or password"}
	ErrUnauthorized          = &DomainError{Code: CodeUnauthorized, Message: "Authentication is required"}
	ErrForbidden             = &DomainError{Code: CodeForbidden, Message: "Insufficient permissions"}
	ErrUserNotFound          = &DomainError{Code: CodeUserNotFound, Message: "User not found"}
	ErrTeamNotFound          = &DomainError{Code: CodeTeamNotFound, Message: "Team not found"}
	ErrNoTeam                = &DomainError{Code: CodeNoTeam, Message: "User is not a member of a team"}
	ErrAlreadyInTeam         = &DomainError{Code: CodeAlreadyInTeam, Message: "User is already a member of a team"}
	ErrTeamFull              = &DomainError{Code: CodeTeamFull, Message: "Maximum team capacity of 4 members reached"}
	ErrInvalidInviteCode     = &DomainError{Code: CodeInvalidInviteCode, Message: "Invite code is invalid"}
	ErrNotTeamCaptain        = &DomainError{Code: CodeNotTeamCaptain, Message: "Only the team captain can perform this action"}
	ErrCaptainCannotLeave    = &DomainError{Code: CodeCaptainCannotLeave, Message: "Captain must transfer ownership or disband the team"}
	ErrCaptainCannotBeKicked = &DomainError{
		Code:    CodeCaptainCannotBeKicked,
		Message: "Team captain cannot be removed from the team",
	}
	ErrTeamMemberNotFound = &DomainError{Code: CodeTeamMemberNotFound, Message: "User is not a member of this team"}
	ErrMutationsLocked    = &DomainError{Code: CodeMutationsLocked, Message: "Team mutations are locked 1 hour before submission deadline"}
	ErrInvalidURLFormat   = &DomainError{Code: CodeInvalidURLFormat, Message: "Solution URL must use HTTP or HTTPS"}
	ErrMinimumTwoMembers  = &DomainError{Code: CodeMinimumTwoMembers, Message: "At least 2 team members are required"}
	ErrDeadlinePassed     = &DomainError{Code: CodeDeadlinePassed, Message: "Submission deadline has passed"}
	ErrInvalidSecretKey   = &DomainError{Code: CodeInvalidSecretKey, Message: "Jury registration key is invalid"}
	ErrInvalidEvaluation  = &DomainError{Code: CodeInvalidEvaluation, Message: "Evaluation contains an invalid criterion or score"}
	ErrEvaluationLocked   = &DomainError{Code: CodeEvaluationLocked, Message: "Evaluations are locked"}
)
