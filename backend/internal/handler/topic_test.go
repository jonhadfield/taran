package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestTopicHandler_List_WithTopics(t *testing.T) {
	h := &TopicHandler{
		Extractions: &testutil.MockExtractionRepo{
			ListTopicsByUserFn: func(_ context.Context, _ string, _ int) ([]string, error) {
				return []string{"tech", "finance", "sports"}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/topics", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var topics []string
	json.NewDecoder(rec.Body).Decode(&topics)
	if len(topics) != 3 {
		t.Fatalf("len = %d, want 3", len(topics))
	}
	if topics[0] != "tech" {
		t.Errorf("topics[0] = %q, want %q", topics[0], "tech")
	}
}

func TestTopicHandler_List_Empty_ReturnsEmptyArray(t *testing.T) {
	h := &TopicHandler{
		Extractions: &testutil.MockExtractionRepo{
			ListTopicsByUserFn: func(_ context.Context, _ string, _ int) ([]string, error) {
				return nil, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/topics", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Must return [] not null
	body := rec.Body.String()
	if body != "[]\n" {
		t.Errorf("body = %q, want %q", body, "[]\n")
	}
}
