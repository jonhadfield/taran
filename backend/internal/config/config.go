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
	Host               string
	Port               int
	TLSDomain          string
	TLSCertDir         string
	APIKey             string
	AllowedOrigins     []string
	BaseURL            string
	UnsubscribeSecret  string
}

type DatabaseConfig struct {
	URL string
}

type WebhookConfig struct {
	Secret string
}

type LLMConfig struct {
	Provider                 string
	AnthropicKey             string
	AnthropicModel           string
	OpenAIKey                string
	OpenAIModel              string
	OllamaURL                string
	OllamaModel              string
	AutoSelectedOverAnthropic bool
}

type DigestConfig struct {
	// Per-user scheduling — no global cron/timezone needed
}

type EmailConfig struct {
	Domain       string
	ResendAPIKey string
	FromAddress  string
}

func Load() (*Config, error) {
	port := 8080
	portEnv := os.Getenv("TARAN_PORT")
	if portEnv == "" {
		portEnv = os.Getenv("PORT") // Cloud Run convention
	}
	if portEnv != "" {
		p, err := strconv.Atoi(portEnv)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", portEnv, err)
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

	var provider string
	var autoSelectedOverAnthropic bool
	if explicit := os.Getenv("TARAN_LLM_PROVIDER"); explicit != "" {
		// Explicit override — validate as before
		provider = explicit
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
	} else {
		// Auto-detect from available API keys, prefer OpenAI
		hasOpenAI := os.Getenv("TARAN_OPENAI_API_KEY") != ""
		hasAnthropic := os.Getenv("TARAN_ANTHROPIC_API_KEY") != ""
		switch {
		case hasOpenAI:
			provider = "openai"
			autoSelectedOverAnthropic = hasAnthropic
		case hasAnthropic:
			provider = "anthropic"
		default:
			return nil, fmt.Errorf("no LLM API key set: provide TARAN_OPENAI_API_KEY or TARAN_ANTHROPIC_API_KEY (or set TARAN_LLM_PROVIDER=ollama)")
		}
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
			Host:              envOr("TARAN_HOST", "0.0.0.0"),
			Port:              port,
			TLSDomain:         os.Getenv("TARAN_TLS_DOMAIN"),
			TLSCertDir:        envOr("TARAN_TLS_CERT_DIR", "certs"),
			APIKey:            apiKey,
			AllowedOrigins:    allowedOrigins,
			BaseURL:           os.Getenv("TARAN_BASE_URL"),
			UnsubscribeSecret: os.Getenv("TARAN_UNSUBSCRIBE_SECRET"),
		},
		DB: DatabaseConfig{
			URL: dbURL,
		},
		Webhook: WebhookConfig{
			Secret: webhookSecret,
		},
		LLM: LLMConfig{
			Provider:                  provider,
			AnthropicKey:              os.Getenv("TARAN_ANTHROPIC_API_KEY"),
			AnthropicModel:            envOr("TARAN_ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
			OpenAIKey:                 os.Getenv("TARAN_OPENAI_API_KEY"),
			OpenAIModel:              envOr("TARAN_OPENAI_MODEL", "gpt-5-mini"),
			OllamaURL:                envOr("TARAN_OLLAMA_URL", "http://localhost:11434"),
			OllamaModel:              os.Getenv("TARAN_OLLAMA_MODEL"),
			AutoSelectedOverAnthropic: autoSelectedOverAnthropic,
		},
		Digest: DigestConfig{},
		Email: EmailConfig{
			Domain:       emailDomain,
			ResendAPIKey: os.Getenv("TARAN_RESEND_API_KEY"),
			FromAddress:  envOr("TARAN_EMAIL_FROM", "MailBrief <digest@mailbrief.io>"),
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
