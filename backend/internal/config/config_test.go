package config

import "testing"

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("CORS_ORIGIN", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_EXPIRATION", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}

	if cfg.DatabaseURL != "postgres://contact:contact@localhost:5432/contact_db?sslmode=disable" {
		t.Errorf(
			"DatabaseURL = %q, want default value",
			cfg.DatabaseURL,
		)
	}

	if cfg.CORSOrigin != "*" {
		t.Errorf("CORSOrigin = %q, want %q", cfg.CORSOrigin, "*")
	}

	if cfg.JWTSecret != "" {
		t.Errorf("JWTSecret = %q, want empty string", cfg.JWTSecret)
	}

	if cfg.JWTExpiration != "1h" {
		t.Errorf("JWTExpiration = %q, want %q", cfg.JWTExpiration, "1h")
	}
}

func TestLoad_EnvironmentValues(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test_db")
	t.Setenv("CORS_ORIGIN", "https://example.com")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRATION", "2h")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test_db" {
		t.Errorf("DatabaseURL = %q, want configured value", cfg.DatabaseURL)
	}

	if cfg.CORSOrigin != "https://example.com" {
		t.Errorf("CORSOrigin = %q, want %q", cfg.CORSOrigin, "https://example.com")
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "test-secret")
	}

	if cfg.JWTExpiration != "2h" {
		t.Errorf("JWTExpiration = %q, want %q", cfg.JWTExpiration, "2h")
	}
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_CONFIG_VALUE", "configured-value")

	if got := getEnv("TEST_CONFIG_VALUE", "fallback"); got != "configured-value" {
		t.Errorf("getEnv() = %q, want %q", got, "configured-value")
	}

	t.Setenv("TEST_CONFIG_EMPTY", "")

	if got := getEnv("TEST_CONFIG_EMPTY", "fallback"); got != "fallback" {
		t.Errorf("getEnv() = %q, want %q", got, "fallback")
	}

	if got := getEnv("TEST_CONFIG_MISSING", "fallback"); got != "fallback" {
		t.Errorf("getEnv() = %q, want %q", got, "fallback")
	}
}
