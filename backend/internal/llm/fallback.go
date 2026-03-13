package llm

import (
	"context"
	"log/slog"
	"strings"

	"github.com/hadfielj/taran/backend/internal/domain"
)

// FallbackProvider wraps a primary and secondary Provider. On transient error
// from the primary, it retries the call against the secondary provider.
// Non-transient errors (empty responses, JSON parse failures) are returned
// immediately since they would fail on any provider.
type FallbackProvider struct {
	primary   Provider
	secondary Provider
}

// NewFallbackProvider creates a provider that falls back to secondary on
// transient errors from primary.
func NewFallbackProvider(primary, secondary Provider) *FallbackProvider {
	return &FallbackProvider{primary: primary, secondary: secondary}
}

func (f *FallbackProvider) Name() string  { return f.primary.Name() }
func (f *FallbackProvider) Model() string { return f.primary.Model() }

func (f *FallbackProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	result, usage, err := f.primary.TriageEmail(ctx, subject, fromAddress, contentPreview)
	if err != nil && isTransient(err) {
		slog.Warn("LLM primary failed, falling back to secondary",
			"op", "triage", "primary", f.primary.Name(), "secondary", f.secondary.Name(), "error", err)
		return f.secondary.TriageEmail(ctx, subject, fromAddress, contentPreview)
	}
	return result, usage, err
}

func (f *FallbackProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	result, usage, err := f.primary.ExtractEmail(ctx, subject, content, fromAddress)
	if err != nil && isTransient(err) {
		slog.Warn("LLM primary failed, falling back to secondary",
			"op", "extract", "primary", f.primary.Name(), "secondary", f.secondary.Name(), "error", err)
		return f.secondary.ExtractEmail(ctx, subject, content, fromAddress)
	}
	return result, usage, err
}

func (f *FallbackProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	result, usage, err := f.primary.GenerateDigest(ctx, extractions, periodType, opts)
	if err != nil && isTransient(err) {
		slog.Warn("LLM primary failed, falling back to secondary",
			"op", "digest", "primary", f.primary.Name(), "secondary", f.secondary.Name(), "error", err)
		return f.secondary.GenerateDigest(ctx, extractions, periodType, opts)
	}
	return result, usage, err
}

// isTransient returns true if the error is likely transient (network, rate
// limit, server error) and worth retrying on a different provider. Errors
// like empty responses or JSON parse failures would fail on any provider.
func isTransient(err error) bool {
	if err == ErrEmptyResponse {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "parse") || strings.Contains(msg, "unmarshal") {
		return false
	}
	return true
}
