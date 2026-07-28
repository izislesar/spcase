package middleware

import (
	"net/http"
	"strings"
)

var sensitiveAPIPrefixes = [...]string{
	"/api/v1/auth",
	"/api/v1/user",
	"/api/v1/team",
	"/api/v1/jury",
	"/api/v1/admin",
}

// NoStoreSensitiveResponses prevents browsers and intermediaries from storing
// responses from API areas that may contain authentication or private data.
func NoStoreSensitiveResponses(next http.Handler) http.Handler {
	if next == nil {
		panic("no-store middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isSensitiveAPIPath(request.URL.Path) {
			writer.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(writer, request)
	})
}

func isSensitiveAPIPath(path string) bool {
	for _, prefix := range sensitiveAPIPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
