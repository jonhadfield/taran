package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIProvider struct {
	client openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIProvider{
		client: client,
		model:  model,
	}
}

func (p *OpenAIProvider) Name() string  { return "openai" }
func (p *OpenAIProvider) Model() string { return p.model }

func (p *OpenAIProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	userPrompt := buildTriageUserPrompt(subject, fromAddress, contentPreview)

	text, usage, err := retryOnEmpty(ctx, TriageTimeout, "openai triage", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(triageSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxCompletionTokens: openai.Int(100),
		})
		if err != nil {
			return "", nil, fmt.Errorf("openai triage: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result TriageResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse triage result: %w (response: %s)", err, text)
	}

	return &result, usage, nil
}

func (p *OpenAIProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	userPrompt := buildExtractionUserPrompt(subject, content, fromAddress)
	if len(userPrompt) > 50000 {
		userPrompt = userPrompt[:50000] + "\n\n[content truncated]"
	}

	text, usage, err := retryOnEmpty(ctx, ExtractTimeout, "openai extract", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(extractionSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxCompletionTokens: openai.Int(4096),
		})
		if err != nil {
			return "", nil, fmt.Errorf("openai extract: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse extraction result: %w (response: %s)", err, text)
	}

	return &result, usage, nil
}

func (p *OpenAIProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	userPrompt := buildDigestUserPrompt(extractions, periodType, opts)

	text, usage, err := retryOnEmpty(ctx, DigestTimeout, "openai digest", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(digestSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxCompletionTokens: openai.Int(1024),
		})
		if err != nil {
			return "", nil, fmt.Errorf("openai digest: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	})
	if err != nil {
		return nil, nil, err
	}

	text = stripCodeFences(text)

	var result DigestSummary
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse digest result: %w (response: %s)", err, text)
	}

	return &result, usage, nil
}

func completionText(c *openai.ChatCompletion) string {
	if len(c.Choices) > 0 {
		return c.Choices[0].Message.Content
	}
	return ""
}

func completionUsage(c *openai.ChatCompletion) *Usage {
	return &Usage{
		InputTokens:  int(c.Usage.PromptTokens),
		OutputTokens: int(c.Usage.CompletionTokens),
		TotalTokens:  int(c.Usage.TotalTokens),
	}
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
