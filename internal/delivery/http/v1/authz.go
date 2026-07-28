package v1

import (
	"net/http"

	"spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/domain"
)

func requireRole(
	writer http.ResponseWriter,
	request *http.Request,
	role domain.Role,
) (middleware.Principal, bool) {
	principal, ok := requireAuthenticated(writer, request)
	if !ok {
		return middleware.Principal{}, false
	}
	if principal.Role != role {
		writeError(writer, http.StatusForbidden, domain.CodeForbidden, domain.ErrForbidden.Message)
		return middleware.Principal{}, false
	}
	return principal, true
}

func requireAuthenticated(
	writer http.ResponseWriter,
	request *http.Request,
) (middleware.Principal, bool) {
	principal, ok := middleware.PrincipalFromContext(request.Context())
	if !ok {
		writeError(writer, http.StatusUnauthorized, domain.CodeUnauthorized, domain.ErrUnauthorized.Message)
		return middleware.Principal{}, false
	}
	return principal, true
}
