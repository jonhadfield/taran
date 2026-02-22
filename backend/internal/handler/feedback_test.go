package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestFeedbackHandler_Upsert_Success(t *testing.T) {
	var saved *domain.EmailFeedback

	h := &FeedbackHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return &domain.Email{ID: "em-1", UserID: "user-1"}, nil
			},
		},
		Feedback: &testutil.MockFeedbackRepo{
			UpsertFn: func(_ context.Context, fb *domain.EmailFeedback) error {
				saved = fb
				return nil
			},
		},
	}

	body := `{"Rating":"useful"}`
	req := httptest.NewRequest("PUT", "/api/emails/em-1/feedback", strings.NewReader(body))
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Upsert(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if saved == nil {
		t.Fatal("expected Upsert to be called")
	}
	if saved.EmailID != "em-1" {
		t.Errorf("EmailID = %q, want %q", saved.EmailID, "em-1")
	}
	if saved.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", saved.UserID, "user-1")
	}
	if saved.Rating != "useful" {
		t.Errorf("Rating = %q, want %q", saved.Rating, "useful")
	}
	if saved.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestFeedbackHandler_Upsert_InvalidRating(t *testing.T) {
	h := &FeedbackHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return &domain.Email{ID: "em-1", UserID: "user-1"}, nil
			},
		},
		Feedback: &testutil.MockFeedbackRepo{},
	}

	body := `{"Rating":"amazing"}`
	req := httptest.NewRequest("PUT", "/api/emails/em-1/feedback", strings.NewReader(body))
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Upsert(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFeedbackHandler_Get_Exists(t *testing.T) {
	h := &FeedbackHandler{
		Feedback: &testutil.MockFeedbackRepo{
			GetByEmailIDFn: func(_ context.Context, _, emailID string) (*domain.EmailFeedback, error) {
				return &domain.EmailFeedback{
					ID:      "fb-1",
					EmailID: emailID,
					Rating:  "not_useful",
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/emails/em-1/feedback", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var fb domain.EmailFeedback
	json.NewDecoder(rec.Body).Decode(&fb)
	if fb.Rating != "not_useful" {
		t.Errorf("Rating = %q, want %q", fb.Rating, "not_useful")
	}
}

func TestFeedbackHandler_Get_NotFound_ReturnsNullRating(t *testing.T) {
	h := &FeedbackHandler{
		Feedback: &testutil.MockFeedbackRepo{
			GetByEmailIDFn: func(_ context.Context, _, _ string) (*domain.EmailFeedback, error) {
				return nil, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/emails/em-1/feedback", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["Rating"] != nil {
		t.Errorf("Rating = %v, want nil", resp["Rating"])
	}
}

func TestFeedbackHandler_Delete_Success(t *testing.T) {
	var deletedUserID, deletedEmailID string

	h := &FeedbackHandler{
		Feedback: &testutil.MockFeedbackRepo{
			DeleteFn: func(_ context.Context, userID, emailID string) error {
				deletedUserID = userID
				deletedEmailID = emailID
				return nil
			},
		},
	}

	req := httptest.NewRequest("DELETE", "/api/emails/em-1/feedback", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if deletedUserID != "user-1" {
		t.Errorf("userID = %q, want %q", deletedUserID, "user-1")
	}
	if deletedEmailID != "em-1" {
		t.Errorf("emailID = %q, want %q", deletedEmailID, "em-1")
	}
}
