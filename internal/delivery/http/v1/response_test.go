package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"spcase.ru/backend/internal/domain"
)

func TestHandleErrorPreservesExistingBusinessError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/team/join", nil)

	handleError(recorder, request, nil, domain.ErrTeamFull)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	var response ErrorResponse
	decoder := json.NewDecoder(recorder.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != domain.CodeTeamFull {
		t.Fatalf("code = %q, want %q", response.Error.Code, domain.CodeTeamFull)
	}
	if response.Error.Message != domain.ErrTeamFull.Message {
		t.Fatalf("message = %q, want %q", response.Error.Message, domain.ErrTeamFull.Message)
	}
}

func TestHandleErrorReturnsExistingContractForPostgresTimeouts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		code string
	}{
		{name: "statement timeout", code: "57014"},
		{name: "lock timeout", code: "55P03"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
			var logOutput bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logOutput, nil))

			timeoutError := &pgconn.PgError{
				Code:    testCase.code,
				Message: testCase.name,
			}
			handleError(recorder, request, logger, fmt.Errorf("repository query: %w", timeoutError))

			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q, want JSON", contentType)
			}
			var response ErrorResponse
			decoder := json.NewDecoder(recorder.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Error.Code != domain.CodeInternal {
				t.Fatalf("code = %q, want %q", response.Error.Code, domain.CodeInternal)
			}
			if response.Error.Message != domain.ErrInternal.Message {
				t.Fatalf("message = %q, want %q", response.Error.Message, domain.ErrInternal.Message)
			}
			if !strings.Contains(logOutput.String(), testCase.code) {
				t.Fatalf("internal log does not preserve PostgreSQL error code %s: %q", testCase.code, logOutput.String())
			}
		})
	}
}
