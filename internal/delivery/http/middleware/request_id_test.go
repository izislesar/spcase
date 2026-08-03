package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRequestIDGeneratesIDWhenHeaderMissing(t *testing.T) {
	var contextID string
	handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		contextID = RequestIDFromContext(request.Context())
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))

	responseID := recorder.Header().Get(RequestIDHeaderName)
	if _, err := uuid.Parse(responseID); err != nil {
		t.Fatalf("response request ID %q is not a UUID", responseID)
	}
	if contextID != responseID {
		t.Fatalf("context request ID %q does not match response header %q", contextID, responseID)
	}
}

func TestRequestIDAcceptsValidInboundID(t *testing.T) {
	inbound := "abc123_DEF-456.789"
	var contextID string
	handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		contextID = RequestIDFromContext(request.Context())
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
	request.Header.Set(RequestIDHeaderName, inbound)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(RequestIDHeaderName); got != inbound {
		t.Fatalf("response request ID = %q, want %q", got, inbound)
	}
	if contextID != inbound {
		t.Fatalf("context request ID = %q, want %q", contextID, inbound)
	}
}

func TestRequestIDRejectsMalformedInboundIDs(t *testing.T) {
	malformed := []string{
		"short",
		strings.Repeat("a", maxInboundRequestIDLength+1),
		"bad id with spaces",
		"bad\nid",
		"bad\rid",
		"bad\tid",
		"bad;id$(x)",
		"1234567%",
	}

	for _, inbound := range malformed {
		t.Run("", func(t *testing.T) {
			var contextID string
			handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
				contextID = RequestIDFromContext(request.Context())
			}))

			request := httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil)
			request.Header.Set(RequestIDHeaderName, inbound)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			responseID := recorder.Header().Get(RequestIDHeaderName)
			if _, err := uuid.Parse(responseID); err != nil {
				t.Fatalf("inbound %q: response request ID %q is not a fresh UUID", inbound, responseID)
			}
			if contextID != responseID {
				t.Fatalf("inbound %q: context request ID %q does not match response header %q",
					inbound, contextID, responseID)
			}
		})
	}
}

func TestRequestIDFromContextWithoutMiddleware(t *testing.T) {
	if got := RequestIDFromContext(t.Context()); got != "" {
		t.Fatalf("request ID = %q, want empty", got)
	}
	if got := RequestIDFromContext(nil); got != "" {
		t.Fatalf("request ID = %q, want empty", got)
	}
}
