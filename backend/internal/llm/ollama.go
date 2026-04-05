package llm

import (
	"context"
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

func (p *OllamaProvider) call(maxTokens int64) callFn {
	return func(ctx context.Context, systemPrompt, userPrompt string) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxTokens: openai.Int(maxTokens),
		})
		if err != nil {
			return "", nil, fmt.Errorf("ollama: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	}
}

func (p *OllamaProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	return triageEmail(ctx, p.call(100), "ollama", subject, fromAddress, contentPreview)
}

func (p *OllamaProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	return extractEmail(ctx, p.call(4096), "ollama", subject, content, fromAddress)
}

func (p *OllamaProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	return generateDigest(ctx, p.call(1024), "ollama", extractions, periodType, opts)
}
