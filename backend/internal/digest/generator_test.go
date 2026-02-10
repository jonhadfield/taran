package digest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestGenerateForUser_Success(t *testing.T) {
	var storedDigest *domain.Digest

	gen := &Generator{
		Extractions: &testutil.MockExtractionRepo{
			ListByUserAndPeriodFn: func(_ context.Context, _ string, _, _ time.Time) ([]domain.Extraction, error) {
				return []domain.Extraction{
					{ID: "ext-1", EmailID: "em-1", Summary: "Summary 1"},
					{ID: "ext-2", EmailID: "em-2", Summary: "Summary 2"},
				}, nil
			},
		},
		Digests: &testutil.MockDigestRepo{
			CreateFn: func(_ context.Context, d *domain.Digest) error {
				storedDigest = d
				return nil
			},
		},
		Provider: &testutil.MockProvider{
			NameVal:  "test-provider",
			ModelVal: "test-model",
			GenerateDigestFn: func(_ context.Context, _ []domain.Extraction, _ string) (*llm.DigestSummary, *llm.Usage, error) {
				return &llm.DigestSummary{
					Title:      "Daily Digest",
					Summary:    "Two emails today",
					Highlights: []string{"highlight1"},
					TopTopics:  []string{"topic1"},
				}, &llm.Usage{TotalTokens: 50}, nil
			},
		},
	}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	d, err := gen.GenerateForUser(context.Background(), "user-1", "daily", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil digest")
	}
	if d.Title != "Daily Digest" {
		t.Errorf("Title = %q, want %q", d.Title, "Daily Digest")
	}
	if d.EmailCount != 2 {
		t.Errorf("EmailCount = %d, want 2", d.EmailCount)
	}
	if len(d.Items) != 2 {
		t.Errorf("Items count = %d, want 2", len(d.Items))
	}
	if storedDigest == nil {
		t.Fatal("digest not stored")
	}
}

func TestGenerateForUser_NoExtractions(t *testing.T) {
	gen := &Generator{
		Extractions: &testutil.MockExtractionRepo{
			ListByUserAndPeriodFn: func(_ context.Context, _ string, _, _ time.Time) ([]domain.Extraction, error) {
				return nil, nil
			},
		},
	}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	d, err := gen.GenerateForUser(context.Background(), "user-1", "daily", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil digest, got %v", d)
	}
}

func TestGenerateForUser_LLMError(t *testing.T) {
	gen := &Generator{
		Extractions: &testutil.MockExtractionRepo{
			ListByUserAndPeriodFn: func(_ context.Context, _ string, _, _ time.Time) ([]domain.Extraction, error) {
				return []domain.Extraction{{ID: "ext-1"}}, nil
			},
		},
		Provider: &testutil.MockProvider{
			GenerateDigestFn: func(_ context.Context, _ []domain.Extraction, _ string) (*llm.DigestSummary, *llm.Usage, error) {
				return nil, nil, fmt.Errorf("LLM error")
			},
		},
	}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	_, err := gen.GenerateForUser(context.Background(), "user-1", "daily", start, end)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGenerateForUser_StoreError(t *testing.T) {
	gen := &Generator{
		Extractions: &testutil.MockExtractionRepo{
			ListByUserAndPeriodFn: func(_ context.Context, _ string, _, _ time.Time) ([]domain.Extraction, error) {
				return []domain.Extraction{{ID: "ext-1", EmailID: "em-1"}}, nil
			},
		},
		Digests: &testutil.MockDigestRepo{
			CreateFn: func(_ context.Context, _ *domain.Digest) error {
				return fmt.Errorf("db error")
			},
		},
		Provider: &testutil.MockProvider{},
	}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	_, err := gen.GenerateForUser(context.Background(), "user-1", "daily", start, end)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGenerateForUser_ItemsSortOrder(t *testing.T) {
	var storedDigest *domain.Digest

	gen := &Generator{
		Extractions: &testutil.MockExtractionRepo{
			ListByUserAndPeriodFn: func(_ context.Context, _ string, _, _ time.Time) ([]domain.Extraction, error) {
				return []domain.Extraction{
					{ID: "ext-1", EmailID: "em-1"},
					{ID: "ext-2", EmailID: "em-2"},
					{ID: "ext-3", EmailID: "em-3"},
				}, nil
			},
		},
		Digests: &testutil.MockDigestRepo{
			CreateFn: func(_ context.Context, d *domain.Digest) error {
				storedDigest = d
				return nil
			},
		},
		Provider: &testutil.MockProvider{},
	}

	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	_, err := gen.GenerateForUser(context.Background(), "user-1", "daily", start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if storedDigest == nil {
		t.Fatal("digest not stored")
	}
	if len(storedDigest.Items) != 3 {
		t.Fatalf("Items count = %d, want 3", len(storedDigest.Items))
	}
	for i, item := range storedDigest.Items {
		if item.SortOrder != i {
			t.Errorf("Items[%d].SortOrder = %d, want %d", i, item.SortOrder, i)
		}
	}
}
