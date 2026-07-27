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
)

const (
	defaultPort      = 8000
	defaultAppDomain = "spcase.ru"
)

// Config contains all runtime settings required by the application.
type Config struct {
	Port                int
	AppDomain           string
	CORSAllowedOrigins  []string
	DB                  DatabaseConfig
	JWT                 JWTConfig
	JuryRegistrationKey string
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
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
	cfg.CORSAllowedOrigins, errs = parseOrigins(getenv("CORS_ALLOWED_ORIGINS"), errs)

	for key, value := range map[string]string{
		"DB_HOST":     cfg.DB.Host,
		"DB_USER":     cfg.DB.User,
		"DB_PASSWORD": cfg.DB.Password,
		"DB_NAME":     cfg.DB.Name,
		"JWT_SECRET":  cfg.JWT.Secret,
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
	if len(errs) > 0 {
		return Config{}, errors.Join(errs...)
	}

	return cfg, nil
}

func parsePort(key, raw string, errs []error) (int, []error) {
	port, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || port < 1 || port > 65535 {
		return 0, append(errs, fmt.Errorf("%s must be an integer between 1 and 65535", key))
	}
	return port, errs
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
