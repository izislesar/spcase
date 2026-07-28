package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spcase.ru/backend/internal/service"
)

type fakePublicService struct{}

func (fakePublicService) Info() service.PublicInfo          { return service.PublicInfo{} }
func (fakePublicService) Schedule() []service.ScheduleEvent { return nil }
func (fakePublicService) FAQ() []service.FAQItem            { return nil }
func (fakePublicService) NoTeam() (string, string)          { return "message", "https://t.me/test" }

type fakePinger struct{ err error }

func (p fakePinger) Ping(context.Context) error { return p.err }

func TestReadinessUsesNestedErrorContract(t *testing.T) {
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: errors.New("offline")})
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
	handler, err := NewPublicHandler(fakePublicService{}, fakePinger{err: errors.New("offline")})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/health/live", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
}
