package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

func (p *AnthropicProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error) {
	userPrompt := buildTriageUserPrompt(subject, fromAddress, contentPreview)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 100,
		System: []anthropic.TextBlockParam{
			{Text: triageSystemPrompt},
		},
		Messages: []anthropic.MessageParam{{
			Role:    anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(userPrompt)},
		}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic triage: %w", err)
	}

	text := extractResponseText(msg)
	if text == "" {
		return nil, nil, fmt.Errorf("empty response from anthropic triage")
	}

	text = stripCodeFences(text)

	var result TriageResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse triage result: %w (response: %s)", err, text)
	}

	usage := &Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
		TotalTokens:  int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
	}

	return &result, usage, nil
}

func (p *AnthropicProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error) {
	userPrompt := buildExtractionUserPrompt(subject, content, fromAddress)

	if len(userPrompt) > 50000 {
		userPrompt = userPrompt[:50000] + "\n\n[content truncated]"
	}

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: extractionSystemPrompt},
		},
		Messages: []anthropic.MessageParam{{
			Role:    anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(userPrompt)},
		}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic extract: %w", err)
	}

	text := extractResponseText(msg)
	if text == "" {
		return nil, nil, fmt.Errorf("empty response from anthropic")
	}

	text = stripCodeFences(text)

	var result ExtractionResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse extraction result: %w (response: %s)", err, text)
	}

	usage := &Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
		TotalTokens:  int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
	}

	return &result, usage, nil
}

func (p *AnthropicProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string) (*DigestSummary, *Usage, error) {
	summaries := make([]string, len(extractions))
	for i, e := range extractions {
		summaries[i] = e.Summary
	}

	userPrompt := buildDigestUserPrompt(summaries, periodType)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: digestSystemPrompt},
		},
		Messages: []anthropic.MessageParam{{
			Role:    anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(userPrompt)},
		}},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic digest: %w", err)
	}

	text := extractResponseText(msg)
	if text == "" {
		return nil, nil, fmt.Errorf("empty response from anthropic")
	}

	text = stripCodeFences(text)

	var result DigestSummary
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, nil, fmt.Errorf("parse digest result: %w (response: %s)", err, text)
	}

	usage := &Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
		TotalTokens:  int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
	}

	return &result, usage, nil
}

func extractResponseText(msg *anthropic.Message) string {
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
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
