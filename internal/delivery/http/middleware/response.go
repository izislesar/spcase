package middleware

import (
	"encoding/json"
	"net/http"

	"spcase.ru/backend/internal/domain"
)

func writeDomainError(writer http.ResponseWriter, status int, domainError *domain.DomainError) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(status)

	response := struct {
		Error struct {
			Code    domain.ErrorCode `json:"code"`
			Message string           `json:"message"`
		} `json:"error"`
	}{}
	response.Error.Code = domainError.Code
	response.Error.Message = domainError.Message
	_ = json.NewEncoder(writer).Encode(response)
}
