package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spcase.ru/backend/internal/domain"
)

func TestAPIErrorResponsesUnknownEndpoint(t *testing.T) {
	handler := newRouterFallbackTestHandler()
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil),
	)

	assertErrorResponse(
		t,
		recorder,
		http.StatusNotFound,
		domain.CodeRouteNotFound,
		domain.ErrRouteNotFound.Message,
	)
}

func TestAPIErrorResponsesWrongMethod(t *testing.T) {
	handler := newRouterFallbackTestHandler()
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodPost, "/api/v1/info", nil),
	)

	assertErrorResponse(
		t,
		recorder,
		http.StatusMethodNotAllowed,
		domain.CodeMethodNotAllowed,
		domain.ErrMethodNotAllowed.Message,
	)
	if allow := recorder.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("Allow = %q, want it to contain %q", allow, http.MethodGet)
	}
}

func TestAPIErrorResponsesPreservesExistingJSONError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/team/my", func(writer http.ResponseWriter, _ *http.Request) {
		writeDomainError(writer, http.StatusNotFound, domain.ErrTeamNotFound)
	})
	recorder := httptest.NewRecorder()

	APIErrorResponses(mux).ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/v1/team/my", nil),
	)

	assertErrorResponse(
		t,
		recorder,
		http.StatusNotFound,
		domain.CodeTeamNotFound,
		domain.ErrTeamNotFound.Message,
	)
}

func TestAPIErrorResponsesLeavesWebErrorsUnchanged(t *testing.T) {
	handler := newRouterFallbackTestHandler()
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/missing-page", nil),
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("Content-Type = %q, want text/plain", contentType)
	}
	if body := recorder.Body.String(); body != "404 page not found\n" {
		t.Fatalf("body = %q", body)
	}
}

func newRouterFallbackTestHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/info", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		http.NotFound(writer, request)
	})
	return APIErrorResponses(mux)
}

func assertErrorResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantStatus int,
	wantCode domain.ErrorCode,
	wantMessage string,
) {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("status = %d, want %d", recorder.Code, wantStatus)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", contentType)
	}

	var response struct {
		Error struct {
			Code    domain.ErrorCode `json:"code"`
			Message string           `json:"message"`
		} `json:"error"`
	}
	decoder := json.NewDecoder(recorder.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("decode response: %v; body = %q", err, recorder.Body.String())
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		t.Fatalf("response contains trailing content: %v; body = %q", err, recorder.Body.String())
	}
	if response.Error.Code != wantCode || response.Error.Message != wantMessage {
		t.Fatalf("error = %#v, want code %q and message %q", response.Error, wantCode, wantMessage)
	}
}
