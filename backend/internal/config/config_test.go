package config

import (
	"strings"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TARAN_DB_URL", "postgres://localhost/test")
	t.Setenv("TARAN_WEBHOOK_SECRET", "test-secret")
	t.Setenv("TARAN_ANTHROPIC_API_KEY", "sk-test-key")
	t.Setenv("TARAN_EMAIL_DOMAIN", "test.example.com")
	t.Setenv("TARAN_API_KEY", "test-api-key")
}

func TestLoad_AllRequired(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DB.URL != "postgres://localhost/test" {
		t.Errorf("DB.URL = %q, want %q", cfg.DB.URL, "postgres://localhost/test")
	}
	if cfg.Webhook.Secret != "test-secret" {
		t.Errorf("Webhook.Secret = %q, want %q", cfg.Webhook.Secret, "test-secret")
	}
	if cfg.Email.Domain != "test.example.com" {
		t.Errorf("Email.Domain = %q, want %q", cfg.Email.Domain, "test.example.com")
	}
}

func TestLoad_MissingDBURL(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_DB_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "TARAN_DB_URL") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "TARAN_DB_URL")
	}
}

func TestLoad_MissingWebhookSecret(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_WEBHOOK_SECRET", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "TARAN_WEBHOOK_SECRET") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "TARAN_WEBHOOK_SECRET")
	}
}

func TestLoad_MissingAnthropicKey(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_ANTHROPIC_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "TARAN_ANTHROPIC_API_KEY") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "TARAN_ANTHROPIC_API_KEY")
	}
}

func TestLoad_MissingEmailDomain(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_EMAIL_DOMAIN", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "TARAN_EMAIL_DOMAIN") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "TARAN_EMAIL_DOMAIN")
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "TARAN_API_KEY") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "TARAN_API_KEY")
	}
}

func TestLoad_AllowedOrigins(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_ALLOWED_ORIGINS", "https://example.com, https://www.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Server.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins len = %d, want 2", len(cfg.Server.AllowedOrigins))
	}
	if cfg.Server.AllowedOrigins[0] != "https://example.com" {
		t.Errorf("AllowedOrigins[0] = %q, want %q", cfg.Server.AllowedOrigins[0], "https://example.com")
	}
	if cfg.Server.AllowedOrigins[1] != "https://www.example.com" {
		t.Errorf("AllowedOrigins[1] = %q, want %q", cfg.Server.AllowedOrigins[1], "https://www.example.com")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid port") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid port")
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", cfg.LLM.Provider, "anthropic")
	}
	// DigestConfig no longer has Cron/Timezone — scheduling is per-user
}

func TestLoad_CustomPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Server.Port)
	}
}

func TestLoad_OllamaProvider(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_LLM_PROVIDER", "ollama")
	t.Setenv("TARAN_ANTHROPIC_API_KEY", "") // not required for ollama

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LLM.Provider != "ollama" {
		t.Errorf("Provider = %q, want %q", cfg.LLM.Provider, "ollama")
	}
}

func TestLoad_UnsupportedProvider(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TARAN_LLM_PROVIDER", "gemini")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "unsupported")
	}
}

func TestAddr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Host: "localhost", Port: 3000},
	}
	got := cfg.Addr()
	if got != "localhost:3000" {
		t.Errorf("Addr() = %q, want %q", got, "localhost:3000")
	}
}
