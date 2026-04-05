package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type AnthropicProvider struct {
	client anthropic.Client
	model  string
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicProvider{
		client: client,
		model:  model,
	}
}

func (p *AnthropicProvider) Name() string  { return "anthropic" }
func (p *AnthropicProvider) Model() string { return p.model }

func (p *AnthropicProvider) call(maxTokens int64) callFn {
	return func(ctx context.Context, systemPrompt, userPrompt string) (string, *Usage, error) {
		msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(p.model),
			MaxTokens: maxTokens,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Messages: []anthropic.MessageParam{{
				Role:    anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(userPrompt)},
			}},
		})
		if err != nil {
			return "", nil, fmt.Errorf("anthropic: %w", err)
		}
		return extractResponseText(msg), anthropicUsage(msg), nil
	}
}

func (p *AnthropicProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	return triageEmail(ctx, p.call(100), "anthropic", subject, fromAddress, contentPreview)
}

func (p *AnthropicProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	return extractEmail(ctx, p.call(4096), "anthropic", subject, content, fromAddress)
}

func (p *AnthropicProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error) {
	return generateDigest(ctx, p.call(1024), "anthropic", extractions, periodType, opts)
}

func extractResponseText(msg *anthropic.Message) string {
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

func anthropicUsage(msg *anthropic.Message) *Usage {
	return &Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
		TotalTokens:  int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
	}
}
