package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNoStoreSensitiveResponses(t *testing.T) {
	t.Parallel()

	endpoints := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/user/me",
		"/api/v1/team/create",
		"/api/v1/team/join",
		"/api/v1/team/my",
		"/api/v1/team/leave",
		"/api/v1/team/kick",
		"/api/v1/team/transfer-ownership",
		"/api/v1/team/disband",
		"/api/v1/team/submit",
		"/api/v1/jury/register",
		"/api/v1/jury/login",
		"/api/v1/jury/teams",
		"/api/v1/jury/evaluations",
		"/api/v1/admin/stats",
		"/api/v1/admin/export/excel",
		"/api/v1/admin/evaluations/close",
		"/api/v1/admin/evaluations/open",
	}

	handler := NoStoreSensitiveResponses(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write([]byte(`{"status":"unchanged"}` + "\n"))
	}))

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, endpoint, nil))

			if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != "no-store" {
				t.Fatalf("Cache-Control = %q, want no-store", cacheControl)
			}
			if recorder.Code != http.StatusCreated {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q", contentType)
			}
			if body := recorder.Body.String(); body != "{\"status\":\"unchanged\"}\n" {
				t.Fatalf("body = %q", body)
			}
		})
	}
}

func TestNoStoreSensitiveResponsesPreservesXLSXResponse(t *testing.T) {
	t.Parallel()

	const contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	payload := []byte("xlsx payload")
	handler := NoStoreSensitiveResponses(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", contentType)
		writer.Header().Set("Content-Disposition", `attachment; filename="hackathon_results.xlsx"`)
		_, _ = writer.Write(payload)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/v1/admin/export/excel", nil),
	)

	if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cacheControl)
	}
	if actual := recorder.Header().Get("Content-Type"); actual != contentType {
		t.Fatalf("Content-Type = %q", actual)
	}
	if disposition := recorder.Header().Get("Content-Disposition"); disposition != `attachment; filename="hackathon_results.xlsx"` {
		t.Fatalf("Content-Disposition = %q", disposition)
	}
	if body := recorder.Body.Bytes(); string(body) != string(payload) {
		t.Fatalf("body = %q", body)
	}
}

func TestNoStoreSensitiveResponsesDoesNotAffectPublicOrStaticResponses(t *testing.T) {
	t.Parallel()

	handler := NoStoreSensitiveResponses(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/static/css/app.css" {
			writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		writer.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		path         string
		cacheControl string
	}{
		{path: "/api/v1/health/live"},
		{path: "/api/v1/health/ready"},
		{path: "/api/v1/info"},
		{path: "/api/v1/schedule"},
		{path: "/api/v1/faq"},
		{path: "/api/v1/no-team"},
		{path: "/static/css/app.css", cacheControl: "public, max-age=31536000, immutable"},
		{path: "/api/v1/authentication"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != test.cacheControl {
				t.Fatalf("Cache-Control = %q, want %q", cacheControl, test.cacheControl)
			}
		})
	}
}
