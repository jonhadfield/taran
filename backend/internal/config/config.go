package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server      ServerConfig
	DB          DatabaseConfig
	Webhook     WebhookConfig
	LLM         LLMConfig
	Digest      DigestConfig
	Email       EmailConfig
	AdminEmails []string
}

type ServerConfig struct {
	Host           string
	Port           int
	TLSDomain      string
	TLSCertDir     string
	APIKey         string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	URL string
}

type WebhookConfig struct {
	Secret string
}

type LLMConfig struct {
	Provider       string
	AnthropicKey   string
	AnthropicModel string
	OpenAIKey      string
	OpenAIModel    string
	OllamaURL      string
	OllamaModel    string
}

type DigestConfig struct {
	Cron     string
	Timezone string
}

type EmailConfig struct {
	Domain string
}

func Load() (*Config, error) {
	port := 8080
	if v := os.Getenv("TARAN_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid TARAN_PORT: %w", err)
		}
		port = p
	}

	dbURL := os.Getenv("TARAN_DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("TARAN_DB_URL is required")
	}

	webhookSecret := os.Getenv("TARAN_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("TARAN_WEBHOOK_SECRET is required")
	}

	provider := envOr("TARAN_LLM_PROVIDER", "anthropic")
	switch provider {
	case "anthropic":
		if os.Getenv("TARAN_ANTHROPIC_API_KEY") == "" {
			return nil, fmt.Errorf("TARAN_ANTHROPIC_API_KEY is required when provider is anthropic")
		}
	case "openai":
		if os.Getenv("TARAN_OPENAI_API_KEY") == "" {
			return nil, fmt.Errorf("TARAN_OPENAI_API_KEY is required when provider is openai")
		}
	case "ollama":
		// ollama doesn't require an API key
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	apiKey := os.Getenv("TARAN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TARAN_API_KEY is required")
	}

	var allowedOrigins []string
	if v := os.Getenv("TARAN_ALLOWED_ORIGINS"); v != "" {
		for _, origin := range strings.Split(v, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	emailDomain := os.Getenv("TARAN_EMAIL_DOMAIN")
	if emailDomain == "" {
		return nil, fmt.Errorf("TARAN_EMAIL_DOMAIN is required")
	}

	var adminEmails []string
	if v := os.Getenv("TARAN_ADMIN_EMAILS"); v != "" {
		for _, email := range strings.Split(v, ",") {
			if trimmed := strings.TrimSpace(email); trimmed != "" {
				adminEmails = append(adminEmails, strings.ToLower(trimmed))
			}
		}
	}

	return &Config{
		Server: ServerConfig{
			Host:           envOr("TARAN_HOST", "0.0.0.0"),
			Port:           port,
			TLSDomain:      os.Getenv("TARAN_TLS_DOMAIN"),
			TLSCertDir:     envOr("TARAN_TLS_CERT_DIR", "certs"),
			APIKey:         apiKey,
			AllowedOrigins: allowedOrigins,
		},
		DB: DatabaseConfig{
			URL: dbURL,
		},
		Webhook: WebhookConfig{
			Secret: webhookSecret,
		},
		LLM: LLMConfig{
			Provider:       provider,
			AnthropicKey:   os.Getenv("TARAN_ANTHROPIC_API_KEY"),
			AnthropicModel: envOr("TARAN_ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
			OpenAIKey:      os.Getenv("TARAN_OPENAI_API_KEY"),
			OpenAIModel:    envOr("TARAN_OPENAI_MODEL", "gpt-4o"),
			OllamaURL:      envOr("TARAN_OLLAMA_URL", "http://localhost:11434"),
			OllamaModel:    os.Getenv("TARAN_OLLAMA_MODEL"),
		},
		Digest: DigestConfig{
			Cron:     envOr("TARAN_DIGEST_CRON", "0 7 * * *"),
			Timezone: envOr("TARAN_DIGEST_TIMEZONE", "UTC"),
		},
		Email: EmailConfig{
			Domain: emailDomain,
		},
		AdminEmails: adminEmails,
	}, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
