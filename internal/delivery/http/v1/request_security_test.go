package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spcase.ru/backend/internal/domain"
)

func TestDecodeJSONAbuseMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		body        []byte
		wantOK      bool
		wantStatus  int
	}{
		{name: "valid", contentType: "application/json", body: []byte(`{"email":"user@example.com","password":"secret"}`), wantOK: true, wantStatus: http.StatusOK},
		{name: "content type with charset", contentType: "application/json; charset=utf-8", body: []byte(`{"email":"user@example.com","password":"secret"}`), wantOK: true, wantStatus: http.StatusOK},
		{name: "empty", contentType: "application/json", body: nil, wantStatus: http.StatusBadRequest},
		{name: "malformed", contentType: "application/json", body: []byte(`{"email":`), wantStatus: http.StatusBadRequest},
		{name: "wrong content type", contentType: "application/x-www-form-urlencoded", body: []byte(`email=user@example.com`), wantStatus: http.StatusUnsupportedMediaType},
		{name: "missing content type", body: []byte(`{}`), wantStatus: http.StatusUnsupportedMediaType},
		{name: "unknown field", contentType: "application/json", body: []byte(`{"email":"user@example.com","password":"secret","admin":true}`), wantStatus: http.StatusBadRequest},
		{name: "duplicate security field", contentType: "application/json", body: []byte(`{"email":"user@example.com","password":"first","password":"second"}`), wantStatus: http.StatusBadRequest},
		{name: "trailing document", contentType: "application/json", body: []byte(`{"email":"user@example.com","password":"secret"} {}`), wantStatus: http.StatusBadRequest},
		{name: "invalid UTF-8", contentType: "application/json", body: []byte{'{', '"', 'e', 'm', 'a', 'i', 'l', '"', ':', '"', 0xff, '"', '}'}, wantStatus: http.StatusBadRequest},
		{name: "oversized", contentType: "application/json", body: bytes.Repeat([]byte("x"), maximumRequestBodyBytes+1), wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			recorder := httptest.NewRecorder()
			var input LoginRequest
			ok := decodeJSON(recorder, request, &input)
			if ok != test.wantOK {
				t.Fatalf("decodeJSON() = %v, want %v; status=%d body=%s", ok, test.wantOK, recorder.Code, recorder.Body.String())
			}
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if !ok {
				var response ErrorResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if response.Error.Code != domain.CodeInvalidRequest {
					t.Fatalf("error code = %q, want %q", response.Error.Code, domain.CodeInvalidRequest)
				}
			}
		})
	}
}

func TestDecodeJSONRejectsNestedDuplicateKeys(t *testing.T) {
	t.Parallel()

	body := `{"team_id":"00000000-0000-0000-0000-000000000001","scores":[{"criterion_id":1,"score":5,"score":10}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/jury/evaluations", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	var input SaveEvaluationsRequest
	if decodeJSON(recorder, request, &input) {
		t.Fatal("nested duplicate key was accepted")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
