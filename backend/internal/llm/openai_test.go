package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestIsReasoningModel(t *testing.T) {
	reasoning := []string{"gpt-5", "gpt-5-mini", "GPT-5-Mini", "o1", "o3-mini", "o4-mini"}
	plain := []string{"gpt-4.1-mini", "gpt-4o", "gpt-4.1", "chatgpt-4o-latest", ""}

	for _, m := range reasoning {
		if !isReasoningModel(m) {
			t.Errorf("%q should be treated as a reasoning model", m)
		}
	}
	for _, m := range plain {
		if isReasoningModel(m) {
			t.Errorf("%q should not be treated as a reasoning model", m)
		}
	}
}

// captureRequest serves one chat completion and returns the request body.
func captureRequest(t *testing.T, model string, maxTokens int64) map[string]any {
	t.Helper()
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"1","object":"chat.completion","model":"m","choices":[{"index":0,"finish_reason":"stop","message":{"role":"assistant","content":"{}"}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	defer srv.Close()

	p := &OpenAIProvider{
		client: openai.NewClient(option.WithAPIKey("test"), option.WithBaseURL(srv.URL)),
		model:  model,
	}
	if _, _, err := p.call(maxTokens)(context.Background(), "system", "user"); err != nil {
		t.Fatalf("call: %v", err)
	}
	return body
}

// Hidden reasoning is billed against the same cap as the reply, so a budget
// sized for the reply alone can return nothing at all.
func TestOpenAICall_ReasoningModelGetsHeadroomAndMinimalEffort(t *testing.T) {
	body := captureRequest(t, "gpt-5-mini", 100)

	if got := body["reasoning_effort"]; got != "minimal" {
		t.Errorf("reasoning_effort = %v, want minimal", got)
	}
	if got := body["max_completion_tokens"]; got != float64(100+openAIReasoningHeadroom) {
		t.Errorf("max_completion_tokens = %v, want %d", got, 100+openAIReasoningHeadroom)
	}
}

func TestOpenAICall_PlainModelUnchanged(t *testing.T) {
	body := captureRequest(t, "gpt-4.1-mini", 100)

	if _, ok := body["reasoning_effort"]; ok {
		t.Error("reasoning_effort must not be sent to a model that rejects it")
	}
	if got := body["max_completion_tokens"]; got != float64(100) {
		t.Errorf("max_completion_tokens = %v, want 100", got)
	}
}
