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

func TestPreferenceHandler_Get_Success(t *testing.T) {
	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{
			GetFn: func(_ context.Context, userID string) (*domain.UserPreference, error) {
				return &domain.UserPreference{
					UserID:          userID,
					DigestEmail:     true,
					DigestFrequency: "daily",
					DigestHour:      9,
					DigestTimezone:  "America/New_York",
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/preferences", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var pref domain.UserPreference
	json.NewDecoder(rec.Body).Decode(&pref)
	if pref.DigestFrequency != "daily" {
		t.Errorf("DigestFrequency = %q, want %q", pref.DigestFrequency, "daily")
	}
	if pref.DigestHour != 9 {
		t.Errorf("DigestHour = %d, want 9", pref.DigestHour)
	}
}

func TestPreferenceHandler_Update_Success(t *testing.T) {
	var upserted *domain.UserPreference

	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{
			GetFn: func(_ context.Context, userID string) (*domain.UserPreference, error) {
				return &domain.UserPreference{
					UserID:          userID,
					DigestEmail:     true,
					DigestFrequency: "daily",
					DigestHour:      9,
					DigestTimezone:  "UTC",
				}, nil
			},
			UpsertFn: func(_ context.Context, pref *domain.UserPreference) error {
				upserted = pref
				return nil
			},
		},
	}

	body := `{"DigestFrequency":"weekly","DigestHour":14}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if upserted == nil {
		t.Fatal("expected Upsert to be called")
	}
	if upserted.DigestFrequency != "weekly" {
		t.Errorf("DigestFrequency = %q, want %q", upserted.DigestFrequency, "weekly")
	}
	if upserted.DigestHour != 14 {
		t.Errorf("DigestHour = %d, want 14", upserted.DigestHour)
	}
	// Unchanged field should be preserved
	if upserted.DigestTimezone != "UTC" {
		t.Errorf("DigestTimezone = %q, want %q", upserted.DigestTimezone, "UTC")
	}
}

func TestPreferenceHandler_Update_InvalidFrequency(t *testing.T) {
	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{},
	}

	body := `{"DigestFrequency":"monthly"}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "frequency") {
		t.Errorf("error = %q, want mention of frequency", resp.Error)
	}
}

func TestPreferenceHandler_Update_InvalidHour(t *testing.T) {
	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{},
	}

	body := `{"DigestHour":25}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "hour") {
		t.Errorf("error = %q, want mention of hour", resp.Error)
	}
}

func TestPreferenceHandler_Update_InvalidDay(t *testing.T) {
	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{},
	}

	body := `{"DigestDay":7}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "day") {
		t.Errorf("error = %q, want mention of day", resp.Error)
	}
}

func TestPreferenceHandler_Update_InvalidTimezone(t *testing.T) {
	h := &PreferenceHandler{
		Preferences: &testutil.MockPreferenceRepo{},
	}

	body := `{"DigestTimezone":"Not/A/Timezone"}`
	req := httptest.NewRequest("PUT", "/api/preferences", strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Error, "timezone") {
		t.Errorf("error = %q, want mention of timezone", resp.Error)
	}
}
