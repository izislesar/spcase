package middleware

import (
	"net/http"

	"spcase.ru/backend/internal/domain"
)

func RequireRoles(roles ...domain.Role) func(http.Handler) http.Handler {
	allowed := make(map[domain.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		if next == nil {
			panic("role middleware next handler cannot be nil")
		}
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			principal, ok := PrincipalFromContext(request.Context())
			if !ok {
				writeUnauthorized(writer)
				return
			}
			if _, ok := allowed[principal.Role]; !ok {
				writeDomainError(writer, http.StatusForbidden, domain.ErrForbidden)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
