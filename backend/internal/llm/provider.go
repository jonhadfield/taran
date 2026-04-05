package llm

import (
	"context"
	"errors"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
)

// ErrEmptyResponse is returned when the LLM API succeeds but returns no content.
var ErrEmptyResponse = errors.New("empty response from LLM")

const (
	TriageTimeout  = 30 * time.Second
	ExtractTimeout = 90 * time.Second
	DigestTimeout  = 90 * time.Second
)

type ExtractionResult struct {
	Summary        string        `json:"summary"`
	KeyPoints      []string      `json:"key_points"`
	Topics         []string      `json:"topics"`
	Links          []domain.Link `json:"links"`
	ActionItems    []string      `json:"action_items"`
	Sentiment      string        `json:"sentiment"`
	SourceCategory string        `json:"source_category"`
}

type DigestSummary struct {
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
	TopTopics  []string `json:"top_topics"`
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type TriageResult struct {
	Extract bool   `json:"extract"`
	Reason  string `json:"reason"`
}

type DigestOptions struct {
	PreferredTopics     []string
	LessPreferredTopics []string
	Style               string // "concise" or "detailed" (default)
	InterestKeywords    []string
	ExclusionKeywords   []string
}

type Provider interface {
	TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*TriageResult, *Usage, error)
	ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*ExtractionResult, *Usage, error)
	GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *DigestOptions) (*DigestSummary, *Usage, error)
	Name() string
	Model() string
}
