package llm

import (
	"context"
	"fmt"

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

func (p *OpenAIProvider) call(maxTokens int64) callFn {
	return func(ctx context.Context, systemPrompt, userPrompt string) (string, *Usage, error) {
		completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxCompletionTokens: openai.Int(maxTokens),
		})
		if err != nil {
			return "", nil, fmt.Errorf("openai: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	}
}

func (p *OpenAIProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	return triageEmail(ctx, p.call(100), "openai", subject, fromAddress, contentPreview)
}

func (p *OpenAIProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	return extractEmail(ctx, p.call(4096), "openai", subject, content, fromAddress)
}

func (p *OpenAIProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	return generateDigest(ctx, p.call(1024), "openai", extractions, periodType, opts)
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
