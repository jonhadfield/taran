package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestProcessor_ProcessEmail_Success(t *testing.T) {
	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{
				ID:          id,
				Subject:     "Newsletter",
				TextBody:    "Content here",
				FromAddress: "sender@example.com",
			}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	extractions := &testutil.MockExtractionRepo{}
	provider := &testutil.MockProvider{
		NameVal:  "test",
		ModelVal: "test-model",
		ExtractEmailFn: func(_ context.Context, subject, content, from string) (*llm.ExtractionResult, *llm.Usage, error) {
			return &llm.ExtractionResult{Summary: "Extracted"}, &llm.Usage{TotalTokens: 50}, nil
		},
	}

	proc := NewProcessor(10, 1, emails, extractions, provider)
	ctx := context.Background()
	proc.Start(ctx)

	proc.Enqueue("email-1")
	proc.Stop()

	// Verify status transitions: processing → processed
	if len(emails.SetStatusCalls) < 2 {
		t.Fatalf("expected at least 2 SetStatus calls, got %d", len(emails.SetStatusCalls))
	}
	first := emails.SetStatusCalls[0]
	if first.ID != "email-1" || first.Status != domain.EmailStatusProcessing {
		t.Errorf("first call = {%q, %q}, want {email-1, processing}", first.ID, first.Status)
	}
	last := emails.SetStatusCalls[len(emails.SetStatusCalls)-1]
	if last.ID != "email-1" || last.Status != domain.EmailStatusProcessed {
		t.Errorf("last call = {%q, %q}, want {email-1, processed}", last.ID, last.Status)
	}
}

func TestProcessor_ProcessEmail_LLMError(t *testing.T) {
	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{ID: id, TextBody: "content"}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	provider := &testutil.MockProvider{
		ExtractEmailFn: func(_ context.Context, _, _, _ string) (*llm.ExtractionResult, *llm.Usage, error) {
			return nil, nil, fmt.Errorf("LLM unavailable")
		},
	}

	proc := NewProcessor(10, 1, emails, &testutil.MockExtractionRepo{}, provider)
	proc.Start(context.Background())

	proc.Enqueue("email-1")
	proc.Stop()

	// Should end with failed status
	found := false
	for _, call := range emails.SetStatusCalls {
		if call.ID == "email-1" && call.Status == domain.EmailStatusFailed {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SetStatus(email-1, failed)")
	}
}

func TestProcessor_ProcessEmail_NoContent(t *testing.T) {
	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{ID: id, TextBody: "", HTMLBody: ""}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}

	proc := NewProcessor(10, 1, emails, &testutil.MockExtractionRepo{}, &testutil.MockProvider{})
	proc.Start(context.Background())

	proc.Enqueue("email-1")
	proc.Stop()

	found := false
	for _, call := range emails.SetStatusCalls {
		if call.ID == "email-1" && call.Status == domain.EmailStatusFailed {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SetStatus(email-1, failed) for empty content")
	}
}

func TestProcessor_ProcessEmail_HTMLFallback(t *testing.T) {
	var gotContent string

	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{
				ID:       id,
				TextBody: "",
				HTMLBody: "<p>HTML content</p>",
			}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	provider := &testutil.MockProvider{
		ExtractEmailFn: func(_ context.Context, _, content, _ string) (*llm.ExtractionResult, *llm.Usage, error) {
			gotContent = content
			return &llm.ExtractionResult{Summary: "test"}, &llm.Usage{TotalTokens: 10}, nil
		},
	}

	proc := NewProcessor(10, 1, emails, &testutil.MockExtractionRepo{}, provider)
	proc.Start(context.Background())

	proc.Enqueue("email-1")
	proc.Stop()

	if gotContent == "" {
		t.Error("expected LLM to receive converted HTML content")
	}
	if gotContent == "<p>HTML content</p>" {
		t.Error("expected HTML to be converted to plain text, got raw HTML")
	}
}

func TestProcessor_Enqueue_QueueFull(t *testing.T) {
	emails := &testutil.MockEmailRepo{
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}

	// Buffer of 1, don't start workers so queue fills up
	proc := NewProcessor(1, 1, emails, &testutil.MockExtractionRepo{}, &testutil.MockProvider{})
	// Don't call Start — queue will never drain

	proc.Enqueue("email-1") // fills the buffer

	// Second enqueue should not block (non-blocking select with default)
	done := make(chan struct{})
	go func() {
		proc.Enqueue("email-2") // should return immediately
		close(done)
	}()

	select {
	case <-done:
		// ok, didn't block
	case <-time.After(time.Second):
		t.Error("Enqueue blocked on full queue")
	}
}

func TestProcessor_ProcessEmail_TriageSkips(t *testing.T) {
	extractCalled := false

	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{
				ID:          id,
				Subject:     "Confirm your subscription",
				TextBody:    "Click here to confirm",
				FromAddress: "noreply@example.com",
			}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	provider := &testutil.MockProvider{
		TriageEmailFn: func(_ context.Context, _, _, _ string) (*llm.TriageResult, *llm.Usage, error) {
			return &llm.TriageResult{Extract: false, Reason: "subscription confirmation"}, &llm.Usage{TotalTokens: 5}, nil
		},
		ExtractEmailFn: func(_ context.Context, _, _, _ string) (*llm.ExtractionResult, *llm.Usage, error) {
			extractCalled = true
			return &llm.ExtractionResult{Summary: "test"}, &llm.Usage{TotalTokens: 10}, nil
		},
	}

	proc := NewProcessor(10, 1, emails, &testutil.MockExtractionRepo{}, provider)
	proc.Start(context.Background())

	proc.Enqueue("email-1")
	proc.Stop()

	if extractCalled {
		t.Error("ExtractEmail should not be called when triage skips")
	}

	found := false
	for _, call := range emails.SetStatusCalls {
		if call.ID == "email-1" && call.Status == domain.EmailStatusSkipped {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected SetStatus(email-1, skipped)")
	}
}

func TestProcessor_ProcessEmail_TriageFailsOpen(t *testing.T) {
	extractCalled := false

	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{
				ID:          id,
				Subject:     "Weekly Newsletter",
				TextBody:    "Content here",
				FromAddress: "news@example.com",
			}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	provider := &testutil.MockProvider{
		TriageEmailFn: func(_ context.Context, _, _, _ string) (*llm.TriageResult, *llm.Usage, error) {
			return nil, nil, fmt.Errorf("triage API error")
		},
		ExtractEmailFn: func(_ context.Context, _, _, _ string) (*llm.ExtractionResult, *llm.Usage, error) {
			extractCalled = true
			return &llm.ExtractionResult{Summary: "Extracted"}, &llm.Usage{TotalTokens: 50}, nil
		},
	}

	proc := NewProcessor(10, 1, emails, &testutil.MockExtractionRepo{}, provider)
	proc.Start(context.Background())

	proc.Enqueue("email-1")
	proc.Stop()

	if !extractCalled {
		t.Error("ExtractEmail should still be called when triage fails (fail open)")
	}

	// Should end with processed status, not failed
	last := emails.SetStatusCalls[len(emails.SetStatusCalls)-1]
	if last.ID != "email-1" || last.Status != domain.EmailStatusProcessed {
		t.Errorf("last status = {%q, %q}, want {email-1, processed}", last.ID, last.Status)
	}
}

func TestProcessor_ConcurrentProcessing(t *testing.T) {
	var processedCount atomic.Int32

	emails := &testutil.MockEmailRepo{
		GetByIDInternalFn: func(_ context.Context, id string) (*domain.Email, error) {
			return &domain.Email{ID: id, TextBody: "content"}, nil
		},
		ListPendingFn: func(_ context.Context, _ int) ([]domain.Email, error) {
			return nil, nil
		},
	}
	provider := &testutil.MockProvider{
		ExtractEmailFn: func(_ context.Context, _, _, _ string) (*llm.ExtractionResult, *llm.Usage, error) {
			time.Sleep(10 * time.Millisecond)
			processedCount.Add(1)
			return &llm.ExtractionResult{Summary: "test"}, &llm.Usage{TotalTokens: 10}, nil
		},
	}

	proc := NewProcessor(10, 2, emails, &testutil.MockExtractionRepo{}, provider)
	proc.Start(context.Background())

	for i := 0; i < 6; i++ {
		proc.Enqueue(fmt.Sprintf("email-%d", i))
	}

	proc.Stop()

	if got := processedCount.Load(); got != 6 {
		t.Errorf("processed %d emails, want 6", got)
	}
}
