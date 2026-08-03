package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"spcase.ru/backend/internal/service"
)

type fakePublicService struct{}

func (fakePublicService) Info() service.PublicInfo          { return service.PublicInfo{} }
func (fakePublicService) Schedule() []service.ScheduleEvent { return nil }
func (fakePublicService) FAQ() []service.FAQItem            { return nil }
func (fakePublicService) NoTeam() (string, string)          { return "message", "https://t.me/test" }

type fakePinger struct{ err error }

func (p fakePinger) Ping(context.Context) error { return p.err }

type testLogCapture struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (capture *testLogCapture) Write(payload []byte) (int, error) {
	capture.mutex.Lock()
	defer capture.mutex.Unlock()
	return capture.buffer.Write(payload)
}

func (capture *testLogCapture) contains(fragment string) bool {
	capture.mutex.Lock()
	defer capture.mutex.Unlock()
	return strings.Contains(capture.buffer.String(), fragment)
}

func TestReadinessUsesNestedErrorContract(t *testing.T) {
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: errors.New("offline")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"error":{"code":"NOT_READY"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestLivenessDoesNotPingDatabase(t *testing.T) {
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: errors.New("offline")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestLivenessStableBody(t *testing.T) {
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	var body struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status = %q, want ok", body.Status)
	}
	if body.Timestamp == "" {
		t.Fatal("timestamp is empty")
	}
}

func TestReadinessSuccessUsesStableBody(t *testing.T) {
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if body.Status != "ready" {
		t.Fatalf("status = %q, want ready", body.Status)
	}
}

func TestReadinessDoesNotLeakDatabaseError(t *testing.T) {
	databaseError := errors.New("dial tcp 10.0.0.5:5432: connection refused")
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: databaseError}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if body := recorder.Body.String(); strings.Contains(body, databaseError.Error()) {
		t.Fatalf("readiness response leaks database error: %s", body)
	}
}

func TestReadinessPingContextIsBounded(t *testing.T) {
	var observedDeadline time.Time
	var observedOK bool
	pinger := pingerFunc(func(ctx context.Context) error {
		observedDeadline, observedOK = ctx.Deadline()
		return errors.New("offline")
	})
	handler, err := NewPublicHandler(fakePublicService{}, pinger, nil)
	if err != nil {
		t.Fatal(err)
	}

	started := time.Now()
	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))

	if !observedOK {
		t.Fatal("readiness ping context has no deadline")
	}
	remaining := time.Until(observedDeadline)
	if remaining <= 0 || remaining > readinessTimeout {
		t.Fatalf("readiness deadline %v is outside (0, %v]", remaining, readinessTimeout)
	}
	if observedDeadline.Before(started) {
		t.Fatal("readiness deadline predates the request")
	}
}

type pingerFunc func(context.Context) error

func (fn pingerFunc) Ping(ctx context.Context) error { return fn(ctx) }

func TestReadinessFailureEmitsStructuredEvent(t *testing.T) {
	capture := &testLogCapture{}
	logger := slog.New(slog.NewJSONHandler(capture, nil))
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: errors.New("offline")}, logger)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if !capture.contains(`"event":"database_readiness_failed"`) {
		t.Fatalf("database_readiness_failed event missing: %s", capture.buffer.String())
	}
	if !capture.contains(`"level":"WARN"`) {
		t.Fatalf("readiness failure is not logged at WARN: %s", capture.buffer.String())
	}
}
