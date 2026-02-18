package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestDashboardHandler_Get_Success(t *testing.T) {
	h := &DashboardHandler{
		Emails: &testutil.MockEmailRepo{
			ListFn: func(_ context.Context, _ string, opts domain.ListOptions) ([]domain.Email, int, error) {
				// Unread count query has IsRead=false
				if opts.IsRead != nil && !*opts.IsRead {
					return nil, 3, nil
				}
				// Total processed query has Status filter
				if opts.Status != nil {
					return nil, 100, nil
				}
				// Recent emails query
				return []domain.Email{
					{ID: "em-1", Subject: "First"},
					{ID: "em-2", Subject: "Second"},
				}, 50, nil
			},
			CountByPeriodFn: func(_ context.Context, _ string, _, _ time.Time) (int, error) {
				return 7, nil
			},
			TopSendersFn: func(_ context.Context, _ string, _, _ time.Time, _ int) ([]domain.SenderCount, error) {
				return []domain.SenderCount{
					{FromAddress: "news@example.com", Count: 3},
				}, nil
			},
			CountByStatusFn: func(_ context.Context, _ string) (map[domain.EmailStatus]int, error) {
				return map[domain.EmailStatus]int{
					domain.EmailStatusPending:    2,
					domain.EmailStatusProcessing: 1,
					domain.EmailStatusFailed:     0,
				}, nil
			},
		},
		Digests: &testutil.MockDigestRepo{
			ListFn: func(_ context.Context, _ string, _ domain.ListOptions) ([]domain.Digest, int, error) {
				return []domain.Digest{
					{ID: "d-1", Title: "Digest 1"},
				}, 1, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp DashboardResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Emails) != 2 {
		t.Errorf("Emails len = %d, want 2", len(resp.Emails))
	}
	if resp.EmailTotal != 50 {
		t.Errorf("EmailTotal = %d, want 50", resp.EmailTotal)
	}
	if len(resp.Digests) != 1 {
		t.Errorf("Digests len = %d, want 1", len(resp.Digests))
	}
	if resp.UnreadCount != 3 {
		t.Errorf("UnreadCount = %d, want 3", resp.UnreadCount)
	}
	if resp.Processing.PendingCount != 2 {
		t.Errorf("PendingCount = %d, want 2", resp.Processing.PendingCount)
	}
	if resp.Processing.ProcessingCount != 1 {
		t.Errorf("ProcessingCount = %d, want 1", resp.Processing.ProcessingCount)
	}
}
