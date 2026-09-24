package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

// openAIReasoningHeadroom is added to the output budget for reasoning models.
// Their hidden reasoning is billed as output and counts against the same cap,
// so a budget sized for the answer alone can be spent before any of the answer
// is written — the call then returns empty and is retried for nothing.
const openAIReasoningHeadroom = 1024

// isReasoningModel reports whether the model reasons before answering. The
// GPT-5 family and the o-series do; gpt-4.1 and earlier do not, and reject the
// reasoning_effort parameter.
func isReasoningModel(model string) bool {
	m := strings.ToLower(model)
	for _, prefix := range []string{"gpt-5", "o1", "o3", "o4"} {
		if strings.HasPrefix(m, prefix) {
			return true
		}
	}
	return false
}

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
		params := openai.ChatCompletionNewParams{
			Model: openai.ChatModel(p.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
			MaxCompletionTokens: openai.Int(maxTokens),
		}
		if isReasoningModel(p.model) {
			// These calls classify and summarise; they need no deliberation.
			params.ReasoningEffort = shared.ReasoningEffortMinimal
			params.MaxCompletionTokens = openai.Int(maxTokens + openAIReasoningHeadroom)
		}

		completion, err := p.client.Chat.Completions.New(ctx, params)
		if err != nil {
			return "", nil, fmt.Errorf("openai: %w", err)
		}
		return completionText(completion), completionUsage(completion), nil
	}
}

func (p *OpenAIProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	return triageEmail(ctx, p.call(100), "openai", subject, fromAddress, contentPreview)
}

func (p *OpenAIProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string, opts *ExtractOptions) (*ExtractionResult, *Usage, error) {
	return extractEmail(ctx, p.call(4096), "openai", subject, content, fromAddress, opts)
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
