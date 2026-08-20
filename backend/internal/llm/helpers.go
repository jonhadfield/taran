package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
)

const maxExtractionPromptLen = 50000 // max characters for the extraction user prompt

// callFn is the signature for a provider-specific LLM API call.
// It takes a context, system prompt, and user prompt, and returns the response text, usage, and error.
type callFn func(ctx context.Context, systemPrompt, userPrompt string) (string, *Usage, error)

// triageEmail builds the triage prompt, calls the LLM via fn, and parses the result.
func triageEmail(ctx context.Context, fn callFn, providerLabel string, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	userPrompt := buildTriageUserPrompt(subject, fromAddress, contentPreview)

	text, usage, err := retryOnEmpty(ctx, TriageTimeout, providerLabel+" triage", func(ctx context.Context) (string, *Usage, error) {
		return fn(ctx, triageSystemPrompt, userPrompt)
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result TriageResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, usage, fmt.Errorf("parse triage result: %w (response: %s)", err, text)
	}

	return &result, usage, nil
}

// extractEmail builds the extraction prompt, calls the LLM via fn, and parses the result.
func extractEmail(ctx context.Context, fn callFn, providerLabel string, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	userPrompt := buildExtractionUserPrompt(subject, content, fromAddress)
	if len(userPrompt) > maxExtractionPromptLen {
		userPrompt = userPrompt[:maxExtractionPromptLen] + "\n\n[content truncated]"
	}

	text, usage, err := retryOnEmpty(ctx, ExtractTimeout, providerLabel+" extract", func(ctx context.Context) (string, *Usage, error) {
		return fn(ctx, extractionSystemPrompt, userPrompt)
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, usage, fmt.Errorf("parse extraction result: %w (response: %s)", err, text)
	}

	// The model's link list is derived from untrusted email content, so keep
	// only links pointing at hosts that genuinely appear in the source.
	if before := len(result.Links); before > 0 {
		result.Links = filterLinksToSource(result.Links, content)
		if dropped := before - len(result.Links); dropped > 0 {
			slog.Warn("dropped extracted links not present in source email",
				"provider", providerLabel, "dropped", dropped, "kept", len(result.Links))
		}
	}

	return &result, usage, nil
}

// generateDigest builds the digest prompt, calls the LLM via fn, and parses the result.
func generateDigest(ctx context.Context, fn callFn, providerLabel string, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	userPrompt := buildDigestUserPrompt(extractions, periodType, opts)

	text, usage, err := retryOnEmpty(ctx, DigestTimeout, providerLabel+" digest", func(ctx context.Context) (string, *Usage, error) {
		return fn(ctx, digestSystemPrompt, userPrompt)
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result DigestSummary
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, usage, fmt.Errorf("parse digest result: %w (response: %s)", err, text)
	}

	return &result, usage, nil
}

// stripCodeFences removes markdown code fences (```json ... ```) that LLMs
// sometimes wrap around JSON responses.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence (```json or ```)
		if i := strings.Index(s, "\n"); i != -1 {
			s = s[i+1:]
		}
		// Remove closing fence
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
		s = strings.TrimSpace(s)
	}
	return s
}

// retryOnEmpty calls fn up to 2 times, retrying once after a short delay if
// the API returns an empty response. The timeout applies per attempt.
func retryOnEmpty(parentCtx context.Context, timeout time.Duration, label string, fn func(ctx context.Context) (string, *Usage, error)) (string, *Usage, error) {
	const maxAttempts = 2

	for attempt := range maxAttempts {
		ctx, cancel := context.WithTimeout(parentCtx, timeout)
		text, usage, err := fn(ctx)
		cancel()

		if err != nil {
			return "", nil, err
		}
		if text != "" {
			return text, usage, nil
		}

		if attempt < maxAttempts-1 {
			slog.Warn("empty response from LLM, retrying", "label", label, "attempt", attempt+1)
			time.Sleep(2 * time.Second)
			continue
		}
	}

	return "", nil, fmt.Errorf("%s: %w", label, ErrEmptyResponse)
}
