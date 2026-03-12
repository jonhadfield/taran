package llm

import (
	"context"
	"log/slog"

	"github.com/hadfielj/taran/backend/internal/config"
	"github.com/hadfielj/taran/backend/internal/crypto"
	"github.com/hadfielj/taran/backend/internal/domain"
)

// UserLLMKeyRepository is the subset of the database layer needed by the resolver.
type UserLLMKeyRepository interface {
	GetByUserAndProvider(ctx context.Context, userID, provider string) (*domain.UserLLMKey, error)
	ListByUser(ctx context.Context, userID string) ([]domain.UserLLMKey, error)
}

// ProviderResolver returns the appropriate LLM provider for a user.
// It checks for a user-owned (BYOK) key first, falling back to the platform provider.
type ProviderResolver struct {
	platform Provider
	keyRepo  UserLLMKeyRepository
	enc      *crypto.Encryptor
	cfg      *config.LLMConfig
}

// NewProviderResolver creates a resolver. enc and keyRepo may be nil if
// BYOK is not configured (encryption key not set).
func NewProviderResolver(platform Provider, keyRepo UserLLMKeyRepository, enc *crypto.Encryptor, cfg *config.LLMConfig) *ProviderResolver {
	return &ProviderResolver{
		platform: platform,
		keyRepo:  keyRepo,
		enc:      enc,
		cfg:      cfg,
	}
}

// ResolveForUser returns a BYOK provider for the user if they have an active
// key for a supported provider, otherwise returns the platform provider.
func (r *ProviderResolver) ResolveForUser(ctx context.Context, userID string) (Provider, error) {
	if r.enc == nil || r.keyRepo == nil {
		return r.platform, nil
	}

	// Try each supported provider in preference order
	for _, providerName := range []string{"anthropic", "openai"} {
		key, err := r.keyRepo.GetByUserAndProvider(ctx, userID, providerName)
		if err != nil {
			slog.Warn("failed to look up BYOK key", "userID", userID, "provider", providerName, "error", err)
			continue
		}
		if key == nil || !key.IsActive {
			continue
		}

		apiKey, err := r.enc.Decrypt(key.EncryptedKey, key.KeyNonce)
		if err != nil {
			slog.Error("failed to decrypt BYOK key", "userID", userID, "provider", providerName, "error", err)
			continue
		}

		provider := r.buildProvider(providerName, apiKey, key.Model)
		if provider != nil {
			slog.Debug("using BYOK provider", "userID", userID, "provider", providerName)
			return provider, nil
		}
	}

	return r.platform, nil
}

// Platform returns the platform (default) provider directly.
func (r *ProviderResolver) Platform() Provider {
	return r.platform
}

func (r *ProviderResolver) buildProvider(providerName, apiKey, model string) Provider {
	switch providerName {
	case "anthropic":
		if model == "" {
			model = r.cfg.AnthropicModel
		}
		return NewAnthropicProvider(apiKey, model)
	case "openai":
		if model == "" {
			model = r.cfg.OpenAIModel
		}
		return NewOpenAIProvider(apiKey, model)
	default:
		return nil
	}
}
