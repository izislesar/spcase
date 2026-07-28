package service

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"spcase.ru/backend/internal/domain"
)

type fakeAuthUsers struct {
	created domain.User
	lookup  *domain.User
}

func (f *fakeAuthUsers) Create(_ context.Context, user domain.User) (domain.User, error) {
	user.ID = uuid.New()
	user.AuthVersion = 1
	f.created = user
	return user, nil
}

func (f *fakeAuthUsers) GetByEmail(_ context.Context, _ string) (domain.User, error) {
	if f.lookup != nil {
		return *f.lookup, nil
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (f *fakeAuthUsers) GetAccountProjection(
	_ context.Context,
	id uuid.UUID,
) (domain.AccountProjection, error) {
	return domain.AccountProjection{
		ID: id, Role: f.created.Role, AuthVersion: f.created.AuthVersion,
	}, nil
}

func TestAuthTokenDoesNotContainTeamIdentity(t *testing.T) {
	users := &fakeAuthUsers{}
	service, err := NewAuthService(
		users, "a sufficiently long JWT secret", "jury-secret", time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Register(context.Background(), RegisterInput{
		FullName: "Test User", University: "University", Telegram: "@test",
		Email: "test@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ValidateToken(result.Token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != result.User.ID.String() || claims.AuthVersion != 1 || claims.Role != domain.RoleUser {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestRegisterJuryRejectsWrongKey(t *testing.T) {
	service, err := NewAuthService(
		&fakeAuthUsers{}, "a sufficiently long JWT secret", "jury-secret", time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.RegisterJury(context.Background(), JuryRegisterInput{SecretKey: "wrong"})
	if !errors.Is(err, domain.ErrInvalidSecretKey) {
		t.Fatalf("error = %v", err)
	}
}

func TestGeneralLoginRejectsJuryRole(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	jury := domain.User{
		ID: uuid.New(), Email: "jury@example.com", PasswordHash: string(hash),
		Role: domain.RoleJury, AuthVersion: 1,
	}
	service, err := NewAuthService(
		&fakeAuthUsers{lookup: &jury},
		"a sufficiently long JWT secret", "jury-secret", time.Now().Add(time.Hour),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Login(context.Background(), jury.Email, "password123")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("error = %v", err)
	}
}

type fakeScoreRepository struct {
	saved []domain.Evaluation
}

func (f *fakeScoreRepository) UpsertBatch(
	_ context.Context,
	evaluations []domain.Evaluation,
) ([]domain.Evaluation, error) {
	f.saved = evaluations
	return evaluations, nil
}
func (f *fakeScoreRepository) ListByJuryID(
	context.Context, uuid.UUID,
) ([]domain.Evaluation, error) {
	return nil, nil
}
func (f *fakeScoreRepository) ListByJuryAndTeamID(
	context.Context, uuid.UUID, uuid.UUID,
) ([]domain.Evaluation, error) {
	return nil, nil
}
func (f *fakeScoreRepository) TeamTotal(
	context.Context, uuid.UUID,
) (domain.TeamScoreTotal, error) {
	return domain.TeamScoreTotal{}, nil
}

func TestScoreServiceRequiresSixCriteria(t *testing.T) {
	repository := &fakeScoreRepository{}
	service, err := NewScoreService(repository)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.SaveEvaluations(context.Background(), uuid.New(), uuid.New(), []CriterionScore{
		{CriterionID: 1, Score: 5},
	})
	if !errors.Is(err, domain.ErrInvalidEvaluation) {
		t.Fatalf("error = %v", err)
	}
	if repository.saved != nil {
		t.Fatal("invalid evaluation was persisted")
	}
}

func TestNormalizeSolutionURL(t *testing.T) {
	if got, err := normalizeSolutionURL(" https://example.com/result "); err != nil || got != "https://example.com/result" {
		t.Fatalf("valid URL = %q, %v", got, err)
	}
	if _, err := normalizeSolutionURL("javascript:alert(1)"); !errors.Is(err, domain.ErrInvalidURLFormat) {
		t.Fatalf("invalid URL error = %v", err)
	}
}

type fakeExportRepository struct{}

func (fakeExportRepository) ExportSummary(context.Context) ([]domain.ExportSummaryRow, error) {
	return []domain.ExportSummaryRow{{
		TeamName: "Team", CaptainName: "Captain", TotalMembers: 2,
		SolutionURL: "https://example.com", TotalScore: 42, EvaluatedByCount: 1,
	}}, nil
}

func (fakeExportRepository) ExportDetails(context.Context) ([]domain.ExportDetailRow, error) {
	return []domain.ExportDetailRow{{
		TeamName: "Team", JuryName: "Jury", CriterionID: 1, Score: 7,
	}}, nil
}

func TestExportServiceWritesXLSX(t *testing.T) {
	export, err := NewExportService(fakeExportRepository{})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := export.WriteXLSX(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(output.Bytes(), []byte("PK")) {
		t.Fatalf("output does not have ZIP/XLSX signature")
	}
}
