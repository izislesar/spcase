package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFromEnv(t *testing.T) {
	t.Parallel()

	cfg, err := LoadFromEnv(mapLookup(validEnvironment()))
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}

	if cfg.Port != defaultPort {
		t.Fatalf("Port = %d, want %d", cfg.Port, defaultPort)
	}
	if cfg.AppDomain != defaultAppDomain {
		t.Fatalf("AppDomain = %q, want %q", cfg.AppDomain, defaultAppDomain)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("CORSAllowedOrigins length = %d, want 2", len(cfg.CORSAllowedOrigins))
	}
	if cfg.RegistrationDeadline.Location() != time.UTC {
		t.Fatalf("RegistrationDeadline location = %v, want UTC", cfg.RegistrationDeadline.Location())
	}
	if got, want := cfg.RegistrationDeadline.Format(time.RFC3339), "2026-10-15T15:00:00Z"; got != want {
		t.Fatalf("RegistrationDeadline = %q, want %q", got, want)
	}
	if got, want := cfg.SubmissionDeadline.Format(time.RFC3339), "2026-10-17T21:00:00Z"; got != want {
		t.Fatalf("SubmissionDeadline = %q, want %q", got, want)
	}
	if got, want := cfg.NoTeamTelegramURL, "https://t.me/example_team_search"; got != want {
		t.Fatalf("NoTeamTelegramURL = %q, want %q", got, want)
	}
}

func TestLoadFromEnvRejectsMissingPublicAndJuryConfig(t *testing.T) {
	t.Parallel()

	environment := validEnvironment()
	delete(environment, "JURY_REGISTRATION_KEY")
	delete(environment, "NO_TEAM_TELEGRAM_URL")

	_, err := LoadFromEnv(mapLookup(environment))
	if err == nil {
		t.Fatal("LoadFromEnv() error = nil, want validation error")
	}
	for _, expected := range []string{"JURY_REGISTRATION_KEY is required", "NO_TEAM_TELEGRAM_URL is required"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("error %q does not contain %q", err, expected)
		}
	}
}

func TestLoadFromEnvRejectsDeadlineOrder(t *testing.T) {
	t.Parallel()

	environment := validEnvironment()
	environment["REGISTRATION_DEADLINE"] = "2026-10-17T21:00:00Z"
	environment["SUBMISSION_DEADLINE"] = "2026-10-17T21:00:00Z"

	_, err := LoadFromEnv(mapLookup(environment))
	if err == nil || !strings.Contains(err.Error(), "REGISTRATION_DEADLINE must be before SUBMISSION_DEADLINE") {
		t.Fatalf("LoadFromEnv() error = %v, want deadline-order validation", err)
	}
}

func TestLoadFromEnvRejectsInvalidPublicURL(t *testing.T) {
	t.Parallel()

	environment := validEnvironment()
	environment["NO_TEAM_TELEGRAM_URL"] = "http://t.me/insecure"

	_, err := LoadFromEnv(mapLookup(environment))
	if err == nil || !strings.Contains(err.Error(), "NO_TEAM_TELEGRAM_URL must be a valid HTTPS URL") {
		t.Fatalf("LoadFromEnv() error = %v, want HTTPS URL validation", err)
	}
}

func validEnvironment() map[string]string {
	return map[string]string{
		"DB_HOST":               "127.0.0.1",
		"DB_PORT":               "5432",
		"DB_USER":               "postgres",
		"DB_PASSWORD":           "postgres",
		"DB_NAME":               "spcase",
		"JWT_SECRET":            "test-jwt-secret",
		"JURY_REGISTRATION_KEY": "test-jury-registration-key",
		"CORS_ALLOWED_ORIGINS":  "https://spcase.ru, https://www.spcase.ru, https://spcase.ru",
		"REGISTRATION_DEADLINE": "2026-10-15T18:00:00+03:00",
		"SUBMISSION_DEADLINE":   "2026-10-17T21:00:00Z",
		"NO_TEAM_TELEGRAM_URL":  "https://t.me/example_team_search",
	}
}

func mapLookup(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}
