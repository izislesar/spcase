package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersApplyToHTMLAPIAndStaticResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		contentType  string
		cacheControl string
		body         string
	}{
		{
			name:        "HTML",
			path:        "/",
			contentType: "text/html; charset=utf-8",
			body:        "<!doctype html><title>СПК</title>",
		},
		{
			name:        "API",
			path:        "/api/v1/info",
			contentType: "application/json; charset=utf-8",
			body:        `{"status":"ok"}`,
		},
		{
			name:         "static",
			path:         "/static/js/app.js",
			contentType:  "text/javascript; charset=utf-8",
			cacheControl: "public, max-age=31536000, immutable",
			body:         "window.SPK = {};",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := SecurityHeaders(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", test.contentType)
				if test.cacheControl != "" {
					writer.Header().Set("Cache-Control", test.cacheControl)
				}
				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte(test.body))
			}))

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))

			assertSecurityHeaders(t, recorder.Header())
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if actual := recorder.Header().Get("Content-Type"); actual != test.contentType {
				t.Fatalf("Content-Type = %q, want %q", actual, test.contentType)
			}
			if actual := recorder.Header().Get("Cache-Control"); actual != test.cacheControl {
				t.Fatalf("Cache-Control = %q, want %q", actual, test.cacheControl)
			}
			if actual := recorder.Body.String(); actual != test.body {
				t.Fatalf("body = %q, want %q", actual, test.body)
			}
		})
	}
}

func assertSecurityHeaders(t *testing.T, headers http.Header) {
	t.Helper()

	expected := map[string]string{
		"X-Content-Type-Options": contentTypeOptions,
		"X-Frame-Options":        frameOptions,
		"Referrer-Policy":        referrerPolicy,
		"Permissions-Policy":     permissionsPolicy,
	}
	for name, want := range expected {
		if actual := headers.Get(name); actual != want {
			t.Errorf("%s = %q, want %q", name, actual, want)
		}
	}
}
