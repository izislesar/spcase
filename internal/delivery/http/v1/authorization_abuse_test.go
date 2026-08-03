package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	httpmiddleware "spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

type principalAuthenticator struct {
	principal httpmiddleware.Principal
}

func (a principalAuthenticator) ValidateToken(string) (service.AccessTokenClaims, error) {
	return service.AccessTokenClaims{
		Role: a.principal.Role, AuthVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{Subject: a.principal.UserID.String()},
	}, nil
}

func (a principalAuthenticator) VerifyAccount(
	context.Context,
	service.AccessTokenClaims,
) (domain.AccountProjection, error) {
	return domain.AccountProjection{
		ID: a.principal.UserID, Role: a.principal.Role, AuthVersion: 1,
	}, nil
}

func authenticatedHandler(t *testing.T, principal httpmiddleware.Principal, handler http.Handler) http.Handler {
	t.Helper()
	auth, err := httpmiddleware.NewAuthMiddleware(principalAuthenticator{principal: principal}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return auth.Middleware(handler)
}

func authenticatedRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: httpmiddleware.AccessTokenCookieName, Value: "test-token"})
	return request
}

type recordingScoreUseCases struct {
	juryID uuid.UUID
	teamID uuid.UUID
}

func (r *recordingScoreUseCases) SaveEvaluations(
	_ context.Context,
	juryID, teamID uuid.UUID,
	_ []service.CriterionScore,
) ([]domain.Evaluation, error) {
	r.juryID, r.teamID = juryID, teamID
	return nil, nil
}

func (*recordingScoreUseCases) JuryEvaluations(context.Context, uuid.UUID) ([]domain.Evaluation, error) {
	return nil, nil
}

func (*recordingScoreUseCases) JuryTeamEvaluations(context.Context, uuid.UUID, uuid.UUID) ([]domain.Evaluation, error) {
	return nil, nil
}

func TestScoreHandlerBindsWritesToAuthenticatedJury(t *testing.T) {
	t.Parallel()

	principalID, otherJuryID, teamID := uuid.New(), uuid.New(), uuid.New()
	scores := &recordingScoreUseCases{}
	handler, err := NewScoreHandler(scores, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"team_id":"` + teamID.String() + `","jury_id":"` + otherJuryID.String() + `","scores":[]}`
	request := authenticatedRequest(http.MethodPost, "/api/v1/jury/evaluations", payload)
	recorder := httptest.NewRecorder()
	authenticatedHandler(t, httpmiddleware.Principal{UserID: principalID, Role: domain.RoleJury}, http.HandlerFunc(handler.SaveEvaluations)).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("injected jury_id status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	payload = `{"team_id":"` + teamID.String() + `","scores":[` +
		`{"criterion_id":1,"score":1},{"criterion_id":2,"score":2},` +
		`{"criterion_id":3,"score":3},{"criterion_id":4,"score":4},` +
		`{"criterion_id":5,"score":5},{"criterion_id":6,"score":6}]}`
	request = authenticatedRequest(http.MethodPost, "/api/v1/jury/evaluations", payload)
	recorder = httptest.NewRecorder()
	authenticatedHandler(t, httpmiddleware.Principal{UserID: principalID, Role: domain.RoleJury}, http.HandlerFunc(handler.SaveEvaluations)).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if scores.juryID != principalID || scores.teamID != teamID {
		t.Fatalf("score identity = jury %s team %s, want jury %s team %s", scores.juryID, scores.teamID, principalID, teamID)
	}
}

type abuseSubmissionUseCases struct {
	captainID uuid.UUID
}

func (r *abuseSubmissionUseCases) Submit(
	_ context.Context,
	captainID uuid.UUID,
	solutionURL string,
) (domain.Submission, error) {
	r.captainID = captainID
	return domain.Submission{SolutionURL: solutionURL, UpdatedAt: time.Now().UTC()}, nil
}

func TestSubmissionHandlerBindsWriteToAuthenticatedUser(t *testing.T) {
	t.Parallel()

	principalID := uuid.New()
	submissions := &abuseSubmissionUseCases{}
	handler, err := NewSubmissionHandler(submissions, nil)
	if err != nil {
		t.Fatal(err)
	}
	request := authenticatedRequest(
		http.MethodPost,
		"/api/v1/team/submit?captain_id="+uuid.NewString(),
		`{"solution_url":"https://example.com/solution"}`,
	)
	recorder := httptest.NewRecorder()
	authenticatedHandler(t, httpmiddleware.Principal{UserID: principalID, Role: domain.RoleUser}, http.HandlerFunc(handler.Submit)).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if submissions.captainID != principalID {
		t.Fatalf("captain ID = %s, want authenticated user %s", submissions.captainID, principalID)
	}
}
