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

func TestSenderHandler_List_MergesPreferences(t *testing.T) {
	h := &SenderHandler{
		Emails: &testutil.MockEmailRepo{
			ListSendersFn: func(_ context.Context, _ string) ([]domain.SenderInfo, error) {
				return []domain.SenderInfo{
					{FromAddress: "news@example.com", FromName: "News", EmailCount: 10},
					{FromAddress: "spam@example.com", FromName: "Spam", EmailCount: 5},
				}, nil
			},
		},
		SenderPrefs: &testutil.MockSenderPreferenceRepo{
			ListByUserFn: func(_ context.Context, _ string) ([]domain.SenderPreference, error) {
				return []domain.SenderPreference{
					{FromAddress: "spam@example.com", Status: "muted"},
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/senders", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var senders []domain.SenderInfo
	json.NewDecoder(rec.Body).Decode(&senders)
	if len(senders) != 2 {
		t.Fatalf("len = %d, want 2", len(senders))
	}
	if senders[0].Status != "normal" {
		t.Errorf("senders[0].Status = %q, want %q", senders[0].Status, "normal")
	}
	if senders[1].Status != "muted" {
		t.Errorf("senders[1].Status = %q, want %q", senders[1].Status, "muted")
	}
}

func TestSenderHandler_Update_Success(t *testing.T) {
	var upserted bool

	h := &SenderHandler{
		SenderPrefs: &testutil.MockSenderPreferenceRepo{
			UpsertFn: func(_ context.Context, _ *domain.SenderPreference) error {
				upserted = true
				return nil
			},
		},
	}

	body := `{"FromAddress":"news@example.com","Status":"muted"}`
	req := httptest.NewRequest("PUT", "/api/senders", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !upserted {
		t.Error("expected Upsert to be called")
	}
}

func TestSenderHandler_Update_MissingAddress(t *testing.T) {
	h := &SenderHandler{
		SenderPrefs: &testutil.MockSenderPreferenceRepo{},
	}

	body := `{"FromAddress":"","Status":"muted"}`
	req := httptest.NewRequest("PUT", "/api/senders", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSenderHandler_Update_InvalidStatus(t *testing.T) {
	h := &SenderHandler{
		SenderPrefs: &testutil.MockSenderPreferenceRepo{},
	}

	body := `{"FromAddress":"news@example.com","Status":"deleted"}`
	req := httptest.NewRequest("PUT", "/api/senders", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSenderHandler_Suggestions_FiltersMutedAndNeedsMinimumFeedback(t *testing.T) {
	h := &SenderHandler{
		Emails: &testutil.MockEmailRepo{
			ListSendersFn: func(_ context.Context, _ string) ([]domain.SenderInfo, error) {
				return []domain.SenderInfo{
					{FromAddress: "bad@example.com", FromName: "Bad Sender"},
					{FromAddress: "okay@example.com", FromName: "Okay Sender"},
					{FromAddress: "muted@example.com", FromName: "Muted"},
					{FromAddress: "few@example.com", FromName: "Few"},
				}, nil
			},
		},
		SenderPrefs: &testutil.MockSenderPreferenceRepo{
			ListByUserFn: func(_ context.Context, _ string) ([]domain.SenderPreference, error) {
				return []domain.SenderPreference{
					{FromAddress: "muted@example.com", Status: "muted"},
				}, nil
			},
		},
		Feedback: &testutil.MockFeedbackRepo{
			GetSenderStatsFn: func(_ context.Context, _ string) ([]domain.SenderFeedbackStat, error) {
				return []domain.SenderFeedbackStat{
					// >60% not useful, >=3 total => suggested
					{FromAddress: "bad@example.com", UsefulCount: 1, NotUsefulCount: 4},
					// 50% not useful => NOT suggested (<=60%)
					{FromAddress: "okay@example.com", UsefulCount: 2, NotUsefulCount: 2},
					// Already muted => excluded
					{FromAddress: "muted@example.com", UsefulCount: 0, NotUsefulCount: 5},
					// <3 total => excluded
					{FromAddress: "few@example.com", UsefulCount: 0, NotUsefulCount: 2},
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/senders/suggestions", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Suggestions(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var suggestions []senderSuggestion
	json.NewDecoder(rec.Body).Decode(&suggestions)
	if len(suggestions) != 1 {
		t.Fatalf("len = %d, want 1", len(suggestions))
	}
	if suggestions[0].FromAddress != "bad@example.com" {
		t.Errorf("FromAddress = %q, want %q", suggestions[0].FromAddress, "bad@example.com")
	}
	if suggestions[0].FromName != "Bad Sender" {
		t.Errorf("FromName = %q, want %q", suggestions[0].FromName, "Bad Sender")
	}
	if suggestions[0].TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", suggestions[0].TotalCount)
	}
}
