package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type logCapture struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (capture *logCapture) Write(payload []byte) (int, error) {
	capture.mutex.Lock()
	defer capture.mutex.Unlock()
	return capture.buffer.Write(payload)
}

func (capture *logCapture) events(t *testing.T) []map[string]any {
	t.Helper()
	capture.mutex.Lock()
	defer capture.mutex.Unlock()

	var events []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(capture.buffer.String()), "\n") {
		if line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("log line is not JSON: %q", line)
		}
		events = append(events, event)
	}
	return events
}

func TestRequestLoggingEmitsCompletedEvent(t *testing.T) {
	capture := &logCapture{}
	logger := slog.New(slog.NewJSONHandler(capture, nil))
	logging := NewRequestLogging(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/things/{id}", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
	})

	handler := RequestID(logging.Middleware(mux))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/things/secret-id", nil))

	events := capture.events(t)
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	event := events[0]
	if event["event"] != "http_request_completed" {
		t.Fatalf("event = %v", event["event"])
	}
	if event["route"] != "GET /api/v1/things/{id}" {
		t.Fatalf("route = %v, want route template", event["route"])
	}
	if event["status"] != float64(http.StatusCreated) {
		t.Fatalf("status = %v", event["status"])
	}
	if event["method"] != http.MethodGet {
		t.Fatalf("method = %v", event["method"])
	}
	if event["request_id"] == "" {
		t.Fatal("request_id is empty")
	}
	if raw := capture.buffer.String(); strings.Contains(raw, "secret-id") {
		t.Fatalf("raw path leaked into logs: %s", raw)
	}
}

func TestRequestLoggingMarksUnmatchedRoutes(t *testing.T) {
	capture := &logCapture{}
	logger := slog.New(slog.NewJSONHandler(capture, nil))
	logging := NewRequestLogging(logger)

	handler := RequestID(logging.Middleware(http.NewServeMux()))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/no/such/route", nil))

	events := capture.events(t)
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0]["route"] != unmatchedRoute {
		t.Fatalf("route = %v, want %q", events[0]["route"], unmatchedRoute)
	}
	if events[0]["status"] != float64(http.StatusNotFound) {
		t.Fatalf("status = %v", events[0]["status"])
	}
}

func TestRequestLoggingLogsServerErrorsAtErrorLevel(t *testing.T) {
	capture := &logCapture{}
	logger := slog.New(slog.NewJSONHandler(capture, nil))
	logging := NewRequestLogging(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /broken", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	})

	handler := RequestID(logging.Middleware(mux))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/broken", nil))

	events := capture.events(t)
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0]["level"] != "ERROR" {
		t.Fatalf("level = %v, want ERROR", events[0]["level"])
	}
	if events[0]["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("status = %v", events[0]["status"])
	}
}

func TestRequestLoggingConcurrentAccess(t *testing.T) {
	capture := &logCapture{}
	logger := slog.New(slog.NewJSONHandler(capture, nil))
	logging := NewRequestLogging(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	handler := RequestID(logging.Middleware(mux))

	var workers sync.WaitGroup
	for index := 0; index < 32; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ok", nil))
			if recorder.Code != http.StatusOK {
				t.Errorf("status = %d", recorder.Code)
			}
		}()
	}
	workers.Wait()

	if events := capture.events(t); len(events) != 32 {
		t.Fatalf("events = %d, want 32", len(events))
	}
}
