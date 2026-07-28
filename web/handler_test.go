package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesEveryPage(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	for route := range pageDefinitions {
		t.Run(route, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
				t.Fatalf("content type = %q", contentType)
			}
			if !strings.Contains(recorder.Body.String(), `src="/static/js/app.js?v=`) {
				t.Fatal("page does not include the frontend bundle")
			}
		})
	}
}

func TestHandlerServesAssetsAndRejectsUnknownPage(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	asset := httptest.NewRecorder()
	handler.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil))
	if asset.Code != http.StatusOK {
		t.Fatalf("asset status = %d", asset.Code)
	}
	if cacheControl := asset.Header().Get("Cache-Control"); cacheControl != "public, max-age=31536000, immutable" {
		t.Fatalf("static asset cache policy = %q", cacheControl)
	}

	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", missing.Code)
	}
}

func TestJuryRootRedirectsToWorkspace(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/jury", nil))
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("status = %d", recorder.Code)
	}
	if location := recorder.Header().Get("Location"); location != "/jury/teams" {
		t.Fatalf("location = %q", location)
	}
}

func TestHomePageUsesRestrainedChampionshipHeading(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	body := recorder.Body.String()
	if !strings.Contains(body, "СПК кейс-чемпионат") {
		t.Fatal("home page does not contain the required championship heading")
	}
	if strings.Contains(body, "Решай.") || strings.Contains(body, "Побеждай.") {
		t.Fatal("home page contains a promotional slogan")
	}
}

func TestPagesUseContentVersionedAssets(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	if len(handler.assetVersion) != 12 {
		t.Fatalf("asset version length = %d", len(handler.assetVersion))
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	for _, asset := range []string{
		`href="/static/css/app.css?v=` + handler.assetVersion + `"`,
		`src="/static/js/app.js?v=` + handler.assetVersion + `"`,
	} {
		if !strings.Contains(body, asset) {
			t.Fatalf("home page does not contain versioned asset %q", asset)
		}
	}
}

func TestPageComponentsDoNotInitializeTwice(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}

	for route := range pageDefinitions {
		t.Run(route, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
			if strings.Contains(recorder.Body.String(), `x-init="init()"`) {
				t.Fatal("page calls init explicitly although Alpine invokes the component lifecycle automatically")
			}
		})
	}
}
