package middleware

import "net/http"

const (
	contentTypeOptions = "nosniff"
	frameOptions       = "DENY"
	referrerPolicy     = "strict-origin-when-cross-origin"
	permissionsPolicy  = "camera=(), microphone=(), geolocation=()"
)

// SecurityHeaders applies baseline browser protections to every HTTP response.
func SecurityHeaders(next http.Handler) http.Handler {
	if next == nil {
		panic("security headers middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", contentTypeOptions)
		writer.Header().Set("X-Frame-Options", frameOptions)
		writer.Header().Set("Referrer-Policy", referrerPolicy)
		writer.Header().Set("Permissions-Policy", permissionsPolicy)
		next.ServeHTTP(writer, request)
	})
}
