package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	httpmiddleware "spcase.ru/backend/internal/delivery/http/middleware"
	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

type fakeAdminUseCases struct{}

func (fakeAdminUseCases) Stats(context.Context) (domain.AdminStats, error) {
	return domain.AdminStats{}, nil
}

func (fakeAdminUseCases) CloseEvaluations(
	context.Context,
	uuid.UUID,
) (domain.EvaluationState, error) {
	return domain.EvaluationState{}, nil
}

func (fakeAdminUseCases) OpenEvaluations(
	context.Context,
	uuid.UUID,
) (domain.EvaluationState, error) {
	return domain.EvaluationState{}, nil
}

type fakeExportUseCases struct {
	payload []byte
	err     error
	calls   int
}

func (f *fakeExportUseCases) WriteXLSX(_ context.Context, writer io.Writer) error {
	f.calls++
	if len(f.payload) > 0 {
		if _, err := writer.Write(f.payload); err != nil {
			return err
		}
	}
	return f.err
}

type handlerExportRepository struct{}

func (handlerExportRepository) ExportSummary(context.Context) ([]domain.ExportSummaryRow, error) {
	return []domain.ExportSummaryRow{{
		TeamName:         "Команда",
		CaptainName:      "Капитан",
		CaptainTelegram:  "@captain",
		Members:          "Капитан, Участник",
		TotalMembers:     2,
		SolutionURL:      "https://example.com/solution",
		TotalScore:       42,
		EvaluatedByCount: 1,
	}}, nil
}

func (handlerExportRepository) ExportDetails(context.Context) ([]domain.ExportDetailRow, error) {
	return []domain.ExportDetailRow{{
		TeamName: "Команда", JuryName: "Эксперт", CriterionID: 1, Score: 7,
	}}, nil
}

type exportAuthenticator struct {
	userID uuid.UUID
	role   domain.Role
}

func (a exportAuthenticator) ValidateToken(string) (service.AccessTokenClaims, error) {
	return service.AccessTokenClaims{
		Role: a.role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: a.userID.String(),
		},
	}, nil
}

func (a exportAuthenticator) VerifyAccount(
	context.Context,
	service.AccessTokenClaims,
) (domain.AccountProjection, error) {
	return domain.AccountProjection{
		ID: a.userID, Role: a.role, AuthVersion: 1,
	}, nil
}

func TestExportExcelReturnsCompleteXLSXAfterSuccessfulGeneration(t *testing.T) {
	export, err := service.NewExportService(handlerExportRepository{})
	if err != nil {
		t.Fatalf("new export service: %v", err)
	}
	handler := newProtectedExportHandler(t, domain.RoleAdmin, export)
	recorder := performAuthenticatedExport(handler)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType !=
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	if disposition := recorder.Header().Get("Content-Disposition"); disposition !=
		`attachment; filename="hackathon_results.xlsx"` {
		t.Fatalf("Content-Disposition = %q", disposition)
	}
	if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("PK")) {
		t.Fatal("response does not have ZIP/XLSX signature")
	}

	file, err := excelize.OpenReader(bytes.NewReader(recorder.Body.Bytes()))
	if err != nil {
		t.Fatalf("open XLSX response: %v", err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close XLSX response: %v", err)
		}
	})
	if sheets := file.GetSheetList(); len(sheets) != 2 ||
		sheets[0] != "Сводка" || sheets[1] != "Детализация по жюри" {
		t.Fatalf("sheets = %#v", sheets)
	}
	if value, err := file.GetCellValue("Сводка", "A2"); err != nil {
		t.Fatalf("read summary team: %v", err)
	} else if value != "Команда" {
		t.Fatalf("summary team = %q, want %q", value, "Команда")
	}
}

func TestExportExcelReturnsJSONWithoutPartialXLSXWhenGenerationFails(t *testing.T) {
	export := &fakeExportUseCases{
		payload: []byte("PK\x03\x04partial xlsx"),
		err:     errors.New("generate XLSX"),
	}
	handler := newProtectedExportHandler(t, domain.RoleAdmin, export)
	recorder := performAuthenticatedExport(handler)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType !=
		"application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	if disposition := recorder.Header().Get("Content-Disposition"); disposition != "" {
		t.Fatalf("Content-Disposition = %q, want empty", disposition)
	}
	if bytes.Contains(recorder.Body.Bytes(), export.payload) {
		t.Fatalf("response contains partial XLSX: %q", recorder.Body.Bytes())
	}
	if bytes.HasPrefix(recorder.Body.Bytes(), []byte("PK")) {
		t.Fatalf("response starts with partial XLSX: %q", recorder.Body.Bytes())
	}

	var response ErrorResponse
	decoder := json.NewDecoder(recorder.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != domain.CodeInternal {
		t.Fatalf("error code = %q, want %q", response.Error.Code, domain.CodeInternal)
	}
	if response.Error.Message != domain.ErrInternal.Message {
		t.Fatalf("error message = %q, want %q", response.Error.Message, domain.ErrInternal.Message)
	}
}

func TestExportExcelStillRequiresAdminRole(t *testing.T) {
	t.Run("missing authentication", func(t *testing.T) {
		export := &fakeExportUseCases{payload: []byte("must not be returned")}
		handler := newProtectedExportHandler(t, domain.RoleAdmin, export)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(
			recorder,
			httptest.NewRequest(http.MethodGet, "/api/v1/admin/export/excel", nil),
		)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
		}
		if export.calls != 0 {
			t.Fatalf("export calls = %d, want 0", export.calls)
		}

		var response ErrorResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if response.Error.Code != domain.CodeUnauthorized {
			t.Fatalf("error code = %q, want %q", response.Error.Code, domain.CodeUnauthorized)
		}
	})

	t.Run("non-admin role", func(t *testing.T) {
		export := &fakeExportUseCases{payload: []byte("must not be returned")}
		handler := newProtectedExportHandler(t, domain.RoleUser, export)
		recorder := performAuthenticatedExport(handler)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
		}
		if export.calls != 0 {
			t.Fatalf("export calls = %d, want 0", export.calls)
		}

		var response ErrorResponse
		if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
			t.Fatalf("decode error response: %v", err)
		}
		if response.Error.Code != domain.CodeForbidden {
			t.Fatalf("error code = %q, want %q", response.Error.Code, domain.CodeForbidden)
		}
	})
}

func newProtectedExportHandler(
	t *testing.T,
	role domain.Role,
	export ExportUseCases,
) http.Handler {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	adminHandler, err := NewAdminHandler(fakeAdminUseCases{}, export, logger)
	if err != nil {
		t.Fatalf("new admin handler: %v", err)
	}
	auth, err := httpmiddleware.NewAuthMiddleware(exportAuthenticator{
		userID: uuid.New(),
		role:   role,
	}, logger)
	if err != nil {
		t.Fatalf("new auth middleware: %v", err)
	}
	return auth.Middleware(
		httpmiddleware.RequireRoles(domain.RoleAdmin)(http.HandlerFunc(adminHandler.ExportExcel)),
	)
}

func performAuthenticatedExport(handler http.Handler) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/export/excel", nil)
	request.AddCookie(&http.Cookie{
		Name:  httpmiddleware.AccessTokenCookieName,
		Value: "valid-token",
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}
