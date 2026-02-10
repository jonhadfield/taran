package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
