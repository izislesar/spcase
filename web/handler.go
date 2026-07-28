// Package web serves the browser UI and its embedded static assets.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
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
	Title        string
	AssetVersion string
}

type pageDefinition struct {
	file  string
	title string
}

var pageDefinitions = map[string]pageDefinition{
	"/":              {file: "index.html", title: "СПК — кейс-чемпионат"},
	"/schedule":      {file: "schedule.html", title: "Расписание — СПК"},
	"/no-team":       {file: "no-team.html", title: "Поиск команды — СПК"},
	"/login":         {file: "login.html", title: "Вход — СПК"},
	"/register":      {file: "register.html", title: "Регистрация — СПК"},
	"/dashboard":     {file: "dashboard.html", title: "Кабинет команды — СПК"},
	"/jury/login":    {file: "jury-login.html", title: "Вход для жюри — СПК"},
	"/jury/register": {file: "jury-register.html", title: "Регистрация жюри — СПК"},
	"/jury/teams":    {file: "jury-teams.html", title: "Работы команд — СПК"},
	"/admin":         {file: "admin.html", title: "Панель администратора — СПК"},
}

// Handler renders known pages and serves build-time-generated assets.
type Handler struct {
	pages        map[string]*template.Template
	static       http.Handler
	assetVersion string
}

// NewHandler parses every template at startup so malformed UI artifacts fail fast.
func NewHandler() (*Handler, error) {
	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	assetVersion, err := embeddedAssetVersion()
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		pages:        make(map[string]*template.Template, len(pageDefinitions)),
		static:       http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
		assetVersion: assetVersion,
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
		writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
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
		Title:        pageDefinitions[request.URL.Path].title,
		AssetVersion: h.assetVersion,
	}); err != nil {
		http.Error(writer, "template rendering failed", http.StatusInternalServerError)
	}
}

func embeddedAssetVersion() (string, error) {
	digest := sha256.New()
	for _, file := range []string{"static/css/app.css", "static/js/app.js"} {
		content, err := fs.ReadFile(assets, file)
		if err != nil {
			return "", errors.New("read embedded asset " + file + ": " + err.Error())
		}
		if _, err := digest.Write(content); err != nil {
			return "", errors.New("hash embedded asset " + file + ": " + err.Error())
		}
	}
	return hex.EncodeToString(digest.Sum(nil))[:12], nil
}

func (h *Handler) setSecurityHeaders(writer http.ResponseWriter) {
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("X-Frame-Options", "DENY")
	writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
}
