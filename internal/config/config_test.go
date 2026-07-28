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
	if got, want := cfg.DB.StatementTimeout, defaultDBStatementTimeout; got != want {
		t.Fatalf("DB.StatementTimeout = %s, want %s", got, want)
	}
	if got, want := cfg.DB.LockTimeout, defaultDBLockTimeout; got != want {
		t.Fatalf("DB.LockTimeout = %s, want %s", got, want)
	}
}

func TestLoadFromEnvLoadsDatabaseTimeouts(t *testing.T) {
	t.Parallel()

	environment := validEnvironment()
	environment["DB_STATEMENT_TIMEOUT"] = "45s"
	environment["DB_LOCK_TIMEOUT"] = "2500ms"

	cfg, err := LoadFromEnv(mapLookup(environment))
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if got, want := cfg.DB.StatementTimeout, 45*time.Second; got != want {
		t.Fatalf("DB.StatementTimeout = %s, want %s", got, want)
	}
	if got, want := cfg.DB.LockTimeout, 2500*time.Millisecond; got != want {
		t.Fatalf("DB.LockTimeout = %s, want %s", got, want)
	}
}

func TestLoadFromEnvRejectsInvalidDatabaseTimeouts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "invalid statement timeout", key: "DB_STATEMENT_TIMEOUT", value: "soon"},
		{name: "disabled statement timeout", key: "DB_STATEMENT_TIMEOUT", value: "0"},
		{name: "negative lock timeout", key: "DB_LOCK_TIMEOUT", value: "-1s"},
		{name: "sub-millisecond lock timeout", key: "DB_LOCK_TIMEOUT", value: "500us"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			environment := validEnvironment()
			environment[testCase.key] = testCase.value

			_, err := LoadFromEnv(mapLookup(environment))
			expected := testCase.key + " must be a duration of at least 1ms"
			if err == nil || !strings.Contains(err.Error(), expected) {
				t.Fatalf("LoadFromEnv() error = %v, want error containing %q", err, expected)
			}
		})
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

func TestLoadFromEnvRejectsUnsafeSecrets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		key           string
		value         string
		expectedError string
	}{
		{
			name:          "empty JWT secret",
			key:           "JWT_SECRET",
			expectedError: "JWT_SECRET is required",
		},
		{
			name:          "empty jury secret",
			key:           "JURY_REGISTRATION_KEY",
			expectedError: "JURY_REGISTRATION_KEY is required",
		},
		{
			name:          "short JWT secret",
			key:           "JWT_SECRET",
			value:         "short-jwt-secret",
			expectedError: "JWT_SECRET must be at least 32 characters long",
		},
		{
			name:          "short jury secret",
			key:           "JURY_REGISTRATION_KEY",
			value:         "short-jury-secret",
			expectedError: "JURY_REGISTRATION_KEY must be at least 32 characters long",
		},
		{
			name:          "placeholder JWT secret",
			key:           "JWT_SECRET",
			value:         "change_me_in_production-use-random-bytes",
			expectedError: "JWT_SECRET must not use a default or placeholder value",
		},
		{
			name:          "placeholder jury secret",
			key:           "JURY_REGISTRATION_KEY",
			value:         "replace_me-with-a-random-jury-secret",
			expectedError: "JURY_REGISTRATION_KEY must not use a default or placeholder value",
		},
		{
			name:          "low-diversity JWT secret",
			key:           "JWT_SECRET",
			value:         strings.Repeat("a", 64),
			expectedError: "JWT_SECRET is too weak",
		},
		{
			name:          "repeated jury secret",
			key:           "JURY_REGISTRATION_KEY",
			value:         strings.Repeat("Ab3$xyZ9", 4),
			expectedError: "JURY_REGISTRATION_KEY is too weak",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			environment := validEnvironment()
			environment[testCase.key] = testCase.value

			_, err := LoadFromEnv(mapLookup(environment))
			if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
				t.Fatalf("LoadFromEnv() error = %v, want error containing %q", err, testCase.expectedError)
			}
		})
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
		"JWT_SECRET":            "8f3c0a7e1d9b5f2468ac0e7d3b9f1a5264c8e0b7d3f9a1e5",
		"JURY_REGISTRATION_KEY": "72e9b4d10fa638c5e27b9d04a16f83c59e2d7a40b6f1c835",
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
