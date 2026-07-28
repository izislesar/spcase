package v1

import (
	"context"
	"encoding/json"
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

type recordingSubmissionUseCases struct {
	calls       int
	solutionURL string
}

func (s *recordingSubmissionUseCases) Submit(
	_ context.Context,
	_ uuid.UUID,
	solutionURL string,
) (domain.Submission, error) {
	s.calls++
	s.solutionURL = solutionURL
	return domain.Submission{
		SolutionURL: solutionURL,
		UpdatedAt:   time.Date(2026, time.October, 17, 20, 15, 0, 0, time.UTC),
	}, nil
}

type submissionAuthenticator struct {
	userID uuid.UUID
}

func (a submissionAuthenticator) ValidateToken(string) (service.AccessTokenClaims, error) {
	return service.AccessTokenClaims{
		Role: domain.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: a.userID.String(),
		},
	}, nil
}

func (a submissionAuthenticator) VerifyAccount(
	context.Context,
	service.AccessTokenClaims,
) (domain.AccountProjection, error) {
	return domain.AccountProjection{
		ID: a.userID, Role: domain.RoleUser, AuthVersion: 1,
	}, nil
}

func TestSubmitSolutionAcceptsValidHTTPAndHTTPSURLs(t *testing.T) {
	t.Parallel()

	for _, solutionURL := range []string{
		"http://example.com/result",
		"https://example.com/result",
	} {
		solutionURL := solutionURL
		t.Run(solutionURL, func(t *testing.T) {
			t.Parallel()

			useCases := &recordingSubmissionUseCases{}
			handler := newAuthenticatedSubmissionHandler(t, useCases)
			recorder := performSubmitSolution(handler, solutionURL)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if useCases.calls != 1 || useCases.solutionURL != solutionURL {
				t.Fatalf("service calls = %d, URL = %q", useCases.calls, useCases.solutionURL)
			}

			var response SubmitSolutionResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Status != "submitted" || response.SolutionURL != solutionURL {
				t.Fatalf("response = %#v", response)
			}
		})
	}
}

func TestSubmitSolutionRejectsInvalidURLAtAPIBoundary(t *testing.T) {
	t.Parallel()

	const prefix = "https://example.com/"
	overlongURL := prefix +
		strings.Repeat("a", domain.MaximumSolutionURLLength-len(prefix)+1)
	testCases := []struct {
		name        string
		solutionURL string
	}{
		{name: "unsupported scheme", solutionURL: "ftp://example.com/result"},
		{name: "malformed URL", solutionURL: "https://exa mple.com/result"},
		{name: "too long", solutionURL: overlongURL},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			useCases := &recordingSubmissionUseCases{}
			handler := newAuthenticatedSubmissionHandler(t, useCases)
			recorder := performSubmitSolution(handler, testCase.solutionURL)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"status = %d, want %d; body = %s",
					recorder.Code,
					http.StatusBadRequest,
					recorder.Body.String(),
				)
			}
			if useCases.calls != 0 {
				t.Fatalf("service calls = %d, want 0", useCases.calls)
			}

			var response ErrorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Error.Code != domain.CodeInvalidURLFormat {
				t.Fatalf(
					"error code = %q, want %q",
					response.Error.Code,
					domain.CodeInvalidURLFormat,
				)
			}
		})
	}
}

func newAuthenticatedSubmissionHandler(
	t *testing.T,
	useCases SubmissionUseCases,
) http.Handler {
	t.Helper()

	handler, err := NewSubmissionHandler(useCases, nil)
	if err != nil {
		t.Fatalf("NewSubmissionHandler(): %v", err)
	}
	auth, err := httpmiddleware.NewAuthMiddleware(
		submissionAuthenticator{userID: uuid.New()},
		nil,
	)
	if err != nil {
		t.Fatalf("NewAuthMiddleware(): %v", err)
	}
	return auth.Middleware(http.HandlerFunc(handler.Submit))
}

func performSubmitSolution(handler http.Handler, solutionURL string) *httptest.ResponseRecorder {
	body, err := json.Marshal(SubmitSolutionRequest{SolutionURL: solutionURL})
	if err != nil {
		panic(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/team/submit", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: httpmiddleware.AccessTokenCookieName, Value: "token"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
