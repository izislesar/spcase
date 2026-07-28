// Package web serves the browser UI and its embedded static assets.
package web

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed template static
var assets embed.FS

type pageData struct {
	Title string
}

type pageDefinition struct {
	file  string
	title string
}

var pageDefinitions = map[string]pageDefinition{
	"/":              {file: "index.html", title: "SPCASE — кейс-чемпионат"},
	"/schedule":      {file: "schedule.html", title: "Расписание — SPCASE"},
	"/no-team":       {file: "no-team.html", title: "Поиск команды — SPCASE"},
	"/login":         {file: "login.html", title: "Вход — SPCASE"},
	"/register":      {file: "register.html", title: "Регистрация — SPCASE"},
	"/dashboard":     {file: "dashboard.html", title: "Кабинет команды — SPCASE"},
	"/jury/login":    {file: "jury-login.html", title: "Вход для жюри — SPCASE"},
	"/jury/register": {file: "jury-register.html", title: "Регистрация жюри — SPCASE"},
	"/jury/teams":    {file: "jury-teams.html", title: "Работы команд — SPCASE"},
	"/admin":         {file: "admin.html", title: "Панель администратора — SPCASE"},
}

// Handler renders known pages and serves build-time-generated assets.
type Handler struct {
	pages  map[string]*template.Template
	static http.Handler
}

// NewHandler parses every template at startup so malformed UI artifacts fail fast.
func NewHandler() (*Handler, error) {
	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		pages:  make(map[string]*template.Template, len(pageDefinitions)),
		static: http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
	}
	for route, definition := range pageDefinitions {
		parsed, parseErr := template.ParseFS(
			assets,
			"template/layout.html",
			path.Join("template", definition.file),
		)
		if parseErr != nil {
			return nil, errors.New("parse page " + route + ": " + parseErr.Error())
		}
		handler.pages[route] = parsed
	}
	return handler, nil
}

// ServeHTTP serves exact page routes and embedded static assets.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.setSecurityHeaders(writer)
	if strings.HasPrefix(request.URL.Path, "/static/") {
		writer.Header().Set("Cache-Control", "public, max-age=3600")
		h.static.ServeHTTP(writer, request)
		return
	}
	if request.URL.Path == "/jury" {
		http.Redirect(writer, request, "/jury/teams", http.StatusTemporaryRedirect)
		return
	}
	page, exists := h.pages[request.URL.Path]
	if !exists {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	if err := page.ExecuteTemplate(writer, "layout", pageData{
		Title: pageDefinitions[request.URL.Path].title,
	}); err != nil {
		http.Error(writer, "template rendering failed", http.StatusInternalServerError)
	}
}

func (h *Handler) setSecurityHeaders(writer http.ResponseWriter) {
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("X-Frame-Options", "DENY")
	writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
}
