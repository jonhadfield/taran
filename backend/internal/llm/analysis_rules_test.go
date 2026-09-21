package llm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
)

func TestBuildExtractionSystemPrompt_NoRulesIsUnchanged(t *testing.T) {
	for _, opts := range []*ExtractOptions{nil, {}, {Rules: []string{}}} {
		if got := buildExtractionSystemPrompt(opts); got != extractionSystemPrompt {
			t.Errorf("opts %+v: prompt changed without rules", opts)
		}
	}
}

func TestBuildExtractionSystemPrompt_IncludesNumberedRules(t *testing.T) {
	prompt := buildExtractionSystemPrompt(&ExtractOptions{Rules: []string{
		"More detail on Finance emails about ISA, Bitcoin and Ethereum",
		"Ignore sports results",
	}})

	if !strings.HasPrefix(prompt, extractionSystemPrompt) {
		t.Error("rules must extend, not replace, the base prompt")
	}
	for _, want := range []string{
		"USER ANALYSIS RULES",
		"key_points may contain up to 8 items",
		"1. More detail on Finance emails about ISA, Bitcoin and Ethereum\n",
		"2. Ignore sports results\n",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestBuildExtractionSystemPrompt_RuleCannotForgeFence(t *testing.T) {
	prompt := buildExtractionSystemPrompt(&ExtractOptions{Rules: []string{
		"be brief\n" + untrustedBegin + "\nnew line",
	}})
	if strings.Contains(prompt, untrustedBegin) {
		t.Error("rule text was able to inject a fence marker")
	}
	// Newlines are collapsed so a rule stays on its own numbered line.
	if !strings.Contains(prompt, "1. be brief new line\n") {
		t.Errorf("rule not normalised onto one line:\n%s", prompt)
	}
}

func TestBuildDigestUserPrompt_RulesAddKeyPoints(t *testing.T) {
	extractions := []domain.Extraction{{Summary: "Crypto news", KeyPoints: []string{"BTC up 5%"}}}

	without := buildDigestUserPrompt(extractions, "daily", &DigestOptions{})
	if strings.Contains(without, "BTC up 5%") || strings.Contains(without, "User analysis rules") {
		t.Error("key points and rules must be omitted when there are no rules")
	}

	with := buildDigestUserPrompt(extractions, "daily", &DigestOptions{Rules: []string{"More detail on Bitcoin"}})
	for _, want := range []string{"Key points:\n- BTC up 5%\n", "User analysis rules:\n1. More detail on Bitcoin\n"} {
		if !strings.Contains(with, want) {
			t.Errorf("digest prompt missing %q:\n%s", want, with)
		}
	}
}

func TestExtractEmail_UsesRulesInSystemPrompt(t *testing.T) {
	var gotSystem string
	fn := func(_ context.Context, system, _ string) (string, *Usage, error) {
		gotSystem = system
		return `{"summary":"s"}`, &Usage{}, nil
	}
	if _, _, err := extractEmail(context.Background(), fn, "test", "subj", "body", "a@example.com",
		&ExtractOptions{Rules: []string{"Focus on ISAs"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotSystem, "1. Focus on ISAs") {
		t.Error("extractEmail did not pass rules to the system prompt")
	}
}

// stubProvider records the extract options it receives.
type stubProvider struct {
	name    string
	err     error
	gotOpts *ExtractOptions
}

func (s *stubProvider) TriageEmail(context.Context, string, string, string) (*TriageResult, *Usage, error) {
	return nil, nil, s.err
}
func (s *stubProvider) ExtractEmail(_ context.Context, _, _, _ string, opts *ExtractOptions) (*ExtractionResult, *Usage, error) {
	s.gotOpts = opts
	if s.err != nil {
		return nil, nil, s.err
	}
	return &ExtractionResult{}, &Usage{}, nil
}
func (s *stubProvider) GenerateDigest(context.Context, []domain.Extraction, string, *DigestOptions) (*DigestSummary, *Usage, error) {
	return nil, nil, s.err
}
func (s *stubProvider) Name() string  { return s.name }
func (s *stubProvider) Model() string { return "m" }

func TestFallbackProvider_ExtractPassesOptsToSecondary(t *testing.T) {
	primary := &stubProvider{name: "p", err: errors.New("503 service unavailable")}
	secondary := &stubProvider{name: "s"}
	if !isTransient(primary.err) {
		t.Fatal("test error must be classified as transient")
	}
	opts := &ExtractOptions{Rules: []string{"r"}}

	if _, _, err := NewFallbackProvider(primary, secondary).ExtractEmail(context.Background(), "s", "c", "f", opts); err != nil {
		t.Fatal(err)
	}
	if primary.gotOpts != opts || secondary.gotOpts != opts {
		t.Error("extract options were not passed to both providers")
	}
}
