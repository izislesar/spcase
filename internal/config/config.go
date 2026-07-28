// Package config loads and validates application configuration.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	defaultPort               = 8000
	defaultAppDomain          = "spcase.ru"
	defaultDBStatementTimeout = 15 * time.Second
	defaultDBLockTimeout      = 5 * time.Second
	minSecretLength           = 32
	minSecretUniqueCharacters = 8
)

var (
	secretPlaceholderFragments = []string{
		"changeme",
		"changemeinproduction",
		"defaultsecret",
		"placeholder",
		"replaceme",
		"supersecretjurykeyfromenv",
		"yourjwtsecret",
		"yourjurysecret",
	}
	weakSecretFragments = []string{
		"letmein",
		"password",
		"qwerty",
	}
)

// Config contains all runtime settings required by the application.
type Config struct {
	Port                 int
	AppDomain            string
	CORSAllowedOrigins   []string
	DB                   DatabaseConfig
	JWT                  JWTConfig
	JuryRegistrationKey  string
	RegistrationDeadline time.Time
	SubmissionDeadline   time.Time
	NoTeamTelegramURL    string
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	Host             string
	Port             int
	User             string
	Password         string
	Name             string
	StatementTimeout time.Duration
	LockTimeout      time.Duration
}

// JWTConfig contains the secret used to sign access tokens.
type JWTConfig struct {
	Secret string
}

// Load reads a local .env file when present, then reads the environment and
// validates every required setting. Existing environment variables take
// precedence over values in .env.
func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	return load(os.Getenv)
}

// LoadFromEnv builds configuration with the supplied environment lookup. It is
// intended for tests and callers that provide configuration programmatically.
func LoadFromEnv(getenv func(string) string) (Config, error) {
	if getenv == nil {
		return Config{}, errors.New("environment lookup cannot be nil")
	}

	return load(getenv)
}

func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		AppDomain:           valueOrDefault(getenv("APP_DOMAIN"), defaultAppDomain),
		JuryRegistrationKey: strings.TrimSpace(getenv("JURY_REGISTRATION_KEY")),
		NoTeamTelegramURL:   strings.TrimSpace(getenv("NO_TEAM_TELEGRAM_URL")),
		DB: DatabaseConfig{
			Host:     strings.TrimSpace(getenv("DB_HOST")),
			User:     strings.TrimSpace(getenv("DB_USER")),
			Password: getenv("DB_PASSWORD"),
			Name:     strings.TrimSpace(getenv("DB_NAME")),
		},
		JWT: JWTConfig{Secret: strings.TrimSpace(getenv("JWT_SECRET"))},
	}

	var errs []error
	cfg.Port, errs = parsePort("PORT", valueOrDefault(getenv("PORT"), strconv.Itoa(defaultPort)), errs)
	cfg.DB.Port, errs = parsePort("DB_PORT", getenv("DB_PORT"), errs)
	cfg.DB.StatementTimeout, errs = parsePositiveDuration(
		"DB_STATEMENT_TIMEOUT",
		valueOrDefault(getenv("DB_STATEMENT_TIMEOUT"), defaultDBStatementTimeout.String()),
		errs,
	)
	cfg.DB.LockTimeout, errs = parsePositiveDuration(
		"DB_LOCK_TIMEOUT",
		valueOrDefault(getenv("DB_LOCK_TIMEOUT"), defaultDBLockTimeout.String()),
		errs,
	)
	cfg.CORSAllowedOrigins, errs = parseOrigins(getenv("CORS_ALLOWED_ORIGINS"), errs)
	cfg.RegistrationDeadline, errs = parseTimestamp("REGISTRATION_DEADLINE", getenv("REGISTRATION_DEADLINE"), errs)
	cfg.SubmissionDeadline, errs = parseTimestamp("SUBMISSION_DEADLINE", getenv("SUBMISSION_DEADLINE"), errs)
	cfg.NoTeamTelegramURL, errs = parseHTTPSURL("NO_TEAM_TELEGRAM_URL", cfg.NoTeamTelegramURL, errs)
	errs = validateSecret("JWT_SECRET", cfg.JWT.Secret, errs)
	errs = validateSecret("JURY_REGISTRATION_KEY", cfg.JuryRegistrationKey, errs)

	for key, value := range map[string]string{
		"DB_HOST":     cfg.DB.Host,
		"DB_USER":     cfg.DB.User,
		"DB_PASSWORD": cfg.DB.Password,
		"DB_NAME":     cfg.DB.Name,
	} {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Errorf("%s is required", key))
		}
	}

	if strings.TrimSpace(cfg.AppDomain) == "" {
		errs = append(errs, errors.New("APP_DOMAIN cannot be empty"))
	}
	if strings.ContainsAny(cfg.AppDomain, "/:\\") {
		errs = append(errs, errors.New("APP_DOMAIN must be a domain name without scheme, path, or port"))
	}
	if !cfg.RegistrationDeadline.IsZero() &&
		!cfg.SubmissionDeadline.IsZero() &&
		!cfg.RegistrationDeadline.Before(cfg.SubmissionDeadline) {
		errs = append(errs, errors.New("REGISTRATION_DEADLINE must be before SUBMISSION_DEADLINE"))
	}
	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}

	return cfg, nil
}

func validateSecret(key, value string, errs []error) []error {
	if value == "" {
		return append(errs, fmt.Errorf("%s is required", key))
	}

	characters := []rune(value)
	if len(characters) < minSecretLength {
		errs = append(errs, fmt.Errorf("%s must be at least %d characters long", key, minSecretLength))
	}

	normalized := normalizeSecret(value)
	for _, placeholder := range secretPlaceholderFragments {
		if strings.Contains(normalized, placeholder) {
			errs = append(errs, fmt.Errorf("%s must not use a default or placeholder value", key))
			break
		}
	}

	uniqueCharacters := make(map[rune]struct{}, len(characters))
	for _, character := range characters {
		uniqueCharacters[character] = struct{}{}
	}
	if len(uniqueCharacters) < minSecretUniqueCharacters ||
		containsAny(normalized, weakSecretFragments) ||
		isRepeatedPattern(characters) {
		errs = append(errs, fmt.Errorf("%s is too weak; use a randomly generated value", key))
	}

	return errs
}

func normalizeSecret(value string) string {
	var normalized strings.Builder
	for _, character := range strings.ToLower(value) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			normalized.WriteRune(character)
		}
	}
	return normalized.String()
}

func containsAny(value string, fragments []string) bool {
	for _, fragment := range fragments {
		if strings.Contains(value, fragment) {
			return true
		}
	}
	return false
}

func isRepeatedPattern(characters []rune) bool {
	for patternLength := 1; patternLength <= len(characters)/2; patternLength++ {
		if len(characters)%patternLength != 0 {
			continue
		}
		repeated := true
		for index := patternLength; index < len(characters); index++ {
			if characters[index] != characters[index%patternLength] {
				repeated = false
				break
			}
		}
		if repeated {
			return true
		}
	}
	return false
}

func parsePort(key, raw string, errs []error) (int, []error) {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port < 1 || port > 65535 {
		return 0, append(errs, fmt.Errorf("%s must be an integer between 1 and 65535", key))
	}
	return port, errs
}

func parsePositiveDuration(key, raw string, errs []error) (time.Duration, []error) {
	duration, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || duration < time.Millisecond {
		return 0, append(errs, fmt.Errorf("%s must be a duration of at least 1ms", key))
	}
	return duration, errs
}

func parseOrigins(raw string, errs []error) ([]string, []error) {
	var origins []string
	seen := make(map[string]struct{})
	for _, rawOrigin := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(rawOrigin)
		if origin == "" {
			continue
		}

		parsed, err := url.ParseRequestURI(origin)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" {
			errs = append(errs, fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid HTTPS origin %q", origin))
			continue
		}
		if _, exists := seen[origin]; !exists {
			origins = append(origins, origin)
			seen[origin] = struct{}{}
		}
	}
	if len(origins) == 0 {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS must contain at least one HTTPS origin"))
	}
	return origins, errs
}

func parseTimestamp(key, raw string, errs []error) (time.Time, []error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, append(errs, fmt.Errorf("%s is required", key))
	}
	deadline, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, append(errs, fmt.Errorf("%s must use RFC3339 format", key))
	}
	return deadline.UTC(), errs
}

func parseHTTPSURL(key, raw string, errs []error) (string, []error) {
	if raw == "" {
		return "", append(errs, fmt.Errorf("%s is required", key))
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return "", append(errs, fmt.Errorf("%s must be a valid HTTPS URL", key))
	}
	return parsed.String(), errs
}

func valueOrDefault(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func loadDotEnv(filename string) error {
	file, err := os.Open(filepath.Clean(filename))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("parse %s:%d: expected KEY=VALUE", filename, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set %s from %s:%d: %w", key, filename, lineNumber, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s: %w", filename, err)
	}

	return nil
}
