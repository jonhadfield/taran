package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestEmailHandler_List_Success(t *testing.T) {
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			ListFn: func(_ context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error) {
				return []domain.Email{
					{ID: "em-1", Subject: "First"},
					{ID: "em-2", Subject: "Second"},
				}, 2, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/emails", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp ListResponse[domain.Email]
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Errorf("data length = %d, want 2", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestEmailHandler_List_WithFilters(t *testing.T) {
	var gotOpts domain.ListOptions

	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			ListFn: func(_ context.Context, _ string, opts domain.ListOptions) ([]domain.Email, int, error) {
				gotOpts = opts
				return nil, 0, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/emails?is_read=true&is_starred=false&limit=10&offset=20", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if gotOpts.IsRead == nil || *gotOpts.IsRead != true {
		t.Error("expected IsRead=true")
	}
	if gotOpts.IsStarred == nil || *gotOpts.IsStarred != false {
		t.Error("expected IsStarred=false")
	}
	if gotOpts.Limit != 10 {
		t.Errorf("Limit = %d, want 10", gotOpts.Limit)
	}
	if gotOpts.Offset != 20 {
		t.Errorf("Offset = %d, want 20", gotOpts.Offset)
	}
}

func TestEmailHandler_Get_Success(t *testing.T) {
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, userID, id string) (*domain.Email, error) {
				return &domain.Email{ID: id, UserID: userID, Subject: "Test"}, nil
			},
		},
		Extractions: &testutil.MockExtractionRepo{
			GetByEmailIDFn: func(_ context.Context, emailID string) (*domain.Extraction, error) {
				return &domain.Extraction{ID: "ext-1", EmailID: emailID, Summary: "Summary"}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/emails/em-1", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp EmailResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != "em-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "em-1")
	}
	if resp.Extraction == nil {
		t.Error("expected extraction to be present")
	}
}

func TestEmailHandler_Get_NotFound(t *testing.T) {
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
		},
		Extractions: &testutil.MockExtractionRepo{},
	}

	req := httptest.NewRequest("GET", "/api/emails/em-1", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestEmailHandler_UpdateState_Success(t *testing.T) {
	var gotState domain.EmailState

	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			UpdateStateFn: func(_ context.Context, _, _ string, state domain.EmailState) error {
				gotState = state
				return nil
			},
		},
	}

	body := `{"IsRead": true, "IsStarred": false}`
	req := httptest.NewRequest("PATCH", "/api/emails/em-1", strings.NewReader(body))
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.UpdateState(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotState.IsRead == nil || *gotState.IsRead != true {
		t.Error("expected IsRead=true")
	}
}

func TestEmailHandler_UpdateState_InvalidBody(t *testing.T) {
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{},
	}

	req := httptest.NewRequest("PATCH", "/api/emails/em-1", strings.NewReader("not json"))
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.UpdateState(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type mockProcessor struct {
	enqueued []string
}

func (m *mockProcessor) Enqueue(emailID string) {
	m.enqueued = append(m.enqueued, emailID)
}

func TestEmailHandler_Reprocess_Success(t *testing.T) {
	proc := &mockProcessor{}
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, _, id string) (*domain.Email, error) {
				return &domain.Email{ID: id, Status: domain.EmailStatusFailed}, nil
			},
			SetStatusFn: func(_ context.Context, _ string, _ domain.EmailStatus, _ string) error {
				return nil
			},
		},
		Extractions: &testutil.MockExtractionRepo{},
		Processor:   proc,
	}

	req := httptest.NewRequest("POST", "/api/emails/em-1/reprocess", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Reprocess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(proc.enqueued) != 1 || proc.enqueued[0] != "em-1" {
		t.Errorf("expected em-1 enqueued, got %v", proc.enqueued)
	}
}

func TestEmailHandler_Reprocess_NotFound(t *testing.T) {
	h := &EmailHandler{
		Emails: &testutil.MockEmailRepo{
			GetByIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
		},
		Extractions: &testutil.MockExtractionRepo{},
	}

	req := httptest.NewRequest("POST", "/api/emails/em-1/reprocess", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Reprocess(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestEmailHandler_Reprocess_ProcessedKeepsExtraction(t *testing.T) {
	proc := &mockProcessor{}
	deleted := false
	emails := &testutil.MockEmailRepo{
		GetByIDFn: func(_ context.Context, _, id string) (*domain.Email, error) {
			return &domain.Email{ID: id, Status: domain.EmailStatusProcessed}, nil
		},
	}
	h := &EmailHandler{
		Emails: emails,
		Extractions: &testutil.MockExtractionRepo{
			DeleteByEmailIDScopedFn: func(_ context.Context, _, _ string) error {
				deleted = true
				return nil
			},
		},
		Processor: proc,
	}

	req := httptest.NewRequest("POST", "/api/emails/em-1/reprocess", nil)
	req.SetPathValue("id", "em-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Reprocess(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if deleted {
		t.Error("re-analysis must keep the existing extraction until it is replaced")
	}
	if len(proc.enqueued) != 1 || proc.enqueued[0] != "em-1" {
		t.Errorf("expected em-1 enqueued, got %v", proc.enqueued)
	}
	if len(emails.SetStatusCalls) != 1 || emails.SetStatusCalls[0].Status != domain.EmailStatusPending {
		t.Errorf("expected status reset to pending, got %+v", emails.SetStatusCalls)
	}
}

func TestEmailHandler_Reprocess_InProgressConflicts(t *testing.T) {
	for _, status := range []domain.EmailStatus{domain.EmailStatusPending, domain.EmailStatusProcessing} {
		proc := &mockProcessor{}
		h := &EmailHandler{
			Emails: &testutil.MockEmailRepo{
				GetByIDFn: func(_ context.Context, _, id string) (*domain.Email, error) {
					return &domain.Email{ID: id, Status: status}, nil
				},
			},
			Extractions: &testutil.MockExtractionRepo{},
			Processor:   proc,
		}

		req := httptest.NewRequest("POST", "/api/emails/em-1/reprocess", nil)
		req.SetPathValue("id", "em-1")
		req = req.WithContext(testutil.ContextWithUserID("user-1"))
		rec := httptest.NewRecorder()
		h.Reprocess(rec, req)

		if rec.Code != http.StatusConflict {
			t.Errorf("%s: status = %d, want %d", status, rec.Code, http.StatusConflict)
		}
		if len(proc.enqueued) != 0 {
			t.Errorf("%s: email was enqueued twice", status)
		}
	}
}

func TestEmailHandler_Reanalyse_QueuesRecentProcessedEmails(t *testing.T) {
	proc := &mockProcessor{}
	var gotOpts domain.ListOptions
	emails := &testutil.MockEmailRepo{
		ListFn: func(_ context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error) {
			if userID != "user-1" {
				t.Errorf("listed emails for %q", userID)
			}
			gotOpts = opts
			return []domain.Email{{ID: "em-1"}, {ID: "em-2"}}, 73, nil
		},
	}
	h := &EmailHandler{Emails: emails, Extractions: &testutil.MockExtractionRepo{}, Processor: proc}

	req := httptest.NewRequest("POST", "/api/emails/reanalyse", strings.NewReader(`{"Days":14}`))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	before := time.Now()
	h.Reanalyse(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	var resp map[string]int
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["queued"] != 2 || resp["matched"] != 73 {
		t.Errorf("response = %v, want queued 2, matched 73", resp)
	}
	if gotOpts.Status == nil || *gotOpts.Status != domain.EmailStatusProcessed {
		t.Error("only processed emails should be re-analysed")
	}
	if gotOpts.Limit != maxReanalyseEmails {
		t.Errorf("limit = %d, want %d", gotOpts.Limit, maxReanalyseEmails)
	}
	wantSince := before.AddDate(0, 0, -14)
	if gotOpts.Since == nil || gotOpts.Since.Sub(wantSince).Abs() > time.Minute {
		t.Errorf("since = %v, want about %v", gotOpts.Since, wantSince)
	}
	if strings.Join(proc.enqueued, ",") != "em-1,em-2" {
		t.Errorf("enqueued = %v", proc.enqueued)
	}
}

func TestEmailHandler_Reanalyse_Validation(t *testing.T) {
	for _, body := range []string{`{"Days":-1}`, `{"Days":31}`, `{`} {
		proc := &mockProcessor{}
		h := &EmailHandler{Emails: &testutil.MockEmailRepo{}, Extractions: &testutil.MockExtractionRepo{}, Processor: proc}
		req := httptest.NewRequest("POST", "/api/emails/reanalyse", strings.NewReader(body))
		req = req.WithContext(testutil.ContextWithUserID("user-1"))
		rec := httptest.NewRecorder()
		h.Reanalyse(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body %s: status = %d, want 400", body, rec.Code)
		}
	}
}
