package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

const maximumRequestBodyBytes = 1 << 20

func decodeJSON(writer http.ResponseWriter, request *http.Request, destination any) bool {
	contentType := request.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		writeError(writer, http.StatusUnsupportedMediaType, domain.CodeInvalidRequest, "Content-Type must be application/json")
		return false
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maximumRequestBodyBytes)
	defer request.Body.Close()
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		var maximumBytesError *http.MaxBytesError
		if errors.As(err, &maximumBytesError) {
			writeError(writer, http.StatusRequestEntityTooLarge, domain.CodeInvalidRequest, "Request body is too large")
			return false
		}
		writeError(writer, http.StatusBadRequest, domain.CodeInvalidRequest, "Request body is invalid")
		return false
	}
	if !utf8.Valid(payload) || validateUniqueJSONKeys(payload) != nil {
		writeError(writer, http.StatusBadRequest, domain.CodeInvalidRequest, "Request body is invalid")
		return false
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(writer, http.StatusBadRequest, domain.CodeInvalidRequest, "Request body is invalid")
		return false
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeError(writer, http.StatusBadRequest, domain.CodeInvalidRequest, "Request body must contain one JSON object")
		return false
	}
	return true
}

func validateUniqueJSONKeys(payload []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	if err := walkJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("JSON contains trailing content")
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delimiter {
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("JSON object key is not a string")
			}
			if _, exists := keys[key]; exists {
				return errors.New("JSON object contains duplicate key")
			}
			keys[key] = struct{}{}
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("JSON object is not closed")
		}
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("JSON array is not closed")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code domain.ErrorCode, message string) {
	writeJSON(writer, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func handleError(writer http.ResponseWriter, request *http.Request, logger *slog.Logger, err error) {
	var validationError *service.ValidationError
	if errors.As(err, &validationError) {
		writeError(writer, http.StatusBadRequest, domain.CodeInvalidRequest, validationError.Error())
		return
	}

	var domainError *domain.DomainError
	if errors.As(err, &domainError) {
		switch domainError.Code {
		case domain.CodeInvalidRequest,
			domain.CodeNoTeam,
			domain.CodeCaptainCannotLeave,
			domain.CodeCaptainCannotBeKicked,
			domain.CodeInvalidURLFormat,
			domain.CodeMinimumTwoMembers,
			domain.CodeInvalidEvaluation:
			writeError(writer, http.StatusBadRequest, domainError.Code, domainError.Message)
		case domain.CodeInvalidCredentials, domain.CodeUnauthorized, domain.CodeAccountDisabled:
			writeError(writer, http.StatusUnauthorized, domainError.Code, domainError.Message)
		case domain.CodeForbidden,
			domain.CodeNotTeamCaptain,
			domain.CodeMutationsLocked,
			domain.CodeDeadlinePassed,
			domain.CodeRegistrationClosed,
			domain.CodeInvalidSecretKey,
			domain.CodeEvaluationLocked:
			writeError(writer, http.StatusForbidden, domainError.Code, domainError.Message)
		case domain.CodeUserNotFound,
			domain.CodeTeamNotFound,
			domain.CodeInvalidInviteCode,
			domain.CodeTeamMemberNotFound,
			domain.CodeSubmissionNotFound:
			writeError(writer, http.StatusNotFound, domainError.Code, domainError.Message)
		case domain.CodeEmailAlreadyExists,
			domain.CodeTeamNameAlreadyExists,
			domain.CodeAlreadyInTeam,
			domain.CodeTeamFull:
			writeError(writer, http.StatusConflict, domainError.Code, domainError.Message)
		default:
			logInternalError(logger, request, err)
			writeError(writer, http.StatusInternalServerError, domain.CodeInternal, domain.ErrInternal.Message)
		}
		return
	}

	logInternalError(logger, request, err)
	writeError(writer, http.StatusInternalServerError, domain.CodeInternal, domain.ErrInternal.Message)
}

func logInternalError(logger *slog.Logger, request *http.Request, err error) {
	logger.ErrorContext(
		request.Context(),
		"HTTP handler failed",
		slog.String("method", request.Method),
		slog.String("path", request.URL.Path),
		slog.String("error", err.Error()),
	)
}

func normalizedLogger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.Default()
	}
	return logger
}

func normalizedDomain(domainName string) string {
	if domainName = strings.TrimSpace(domainName); domainName != "" {
		return domainName
	}
	return "spcase.ru"
}
