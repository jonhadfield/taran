package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// OllamaProvider uses Ollama's OpenAI-compatible API endpoint.
type OllamaProvider struct {
	client openai.Client
	model  string
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	client := openai.NewClient(
		option.WithBaseURL(baseURL+"/v1"),
		option.WithAPIKey("ollama"), // Ollama doesn't require a real key
	)
	return &OllamaProvider{
		client: client,
		model:  model,
	}
}

func (p *OllamaProvider) Name() string  { return "ollama" }
func (p *OllamaProvider) Model() string { return p.model }

func (p *OllamaProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	userPrompt := buildTriageUserPrompt(subject, fromAddress, contentPreview)

	text, usage, err := retryOnEmpty(ctx, TriageTimeout, "ollama triage", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(triageSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxTokens: openai.Int(100),
		})
		if err != nil {
			return "", nil, fmt.Errorf("ollama triage: %w", err)
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

func (p *OllamaProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	userPrompt := buildExtractionUserPrompt(subject, content, fromAddress)
	if len(userPrompt) > 50000 {
		userPrompt = userPrompt[:50000] + "\n\n[content truncated]"
	}

	text, usage, err := retryOnEmpty(ctx, ExtractTimeout, "ollama extract", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(extractionSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxTokens: openai.Int(4096),
		})
		if err != nil {
			return "", nil, fmt.Errorf("ollama extract: %w", err)
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

func (p *OllamaProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	userPrompt := buildDigestUserPrompt(extractions, periodType, opts)

	text, usage, err := retryOnEmpty(ctx, DigestTimeout, "ollama digest", func(ctx context.Context) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(digestSystemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxTokens: openai.Int(1024),
		})
		if err != nil {
			return "", nil, fmt.Errorf("ollama digest: %w", err)
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
