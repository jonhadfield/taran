package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

// memSettings is an in-memory AppSettings.
type memSettings map[string]string

func (m memSettings) GetBool(_ context.Context, key string, fallback bool) (bool, error) {
	v, ok := m[key]
	if !ok {
		return fallback, nil
	}
	return v == "true", nil
}

func (m memSettings) GetInt(_ context.Context, _ string, fallback int) (int, error) {
	return fallback, nil
}

func (m memSettings) Set(_ context.Context, key, value string) error {
	m[key] = value
	return nil
}

func TestInviteHandler_CheckAccess_OpenRegistration(t *testing.T) {
	var created *domain.Invite
	var markedAccepted string
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			CreateFn: func(_ context.Context, inv *domain.Invite) error {
				created = inv
				return nil
			},
			MarkAcceptedFn: func(_ context.Context, email string) (bool, error) {
				markedAccepted = email
				return true, nil
			},
		},
		Settings: memSettings{auth.OpenRegistrationSetting: "true"},
	}

	req := httptest.NewRequest("GET", "/api/access", nil)
	req = req.WithContext(contextWithUserEmail("user-9", "New@Example.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if rec.Code != http.StatusOK || resp["hasAccess"] != true || resp["reason"] != auth.AccessOpenRegistration {
		t.Fatalf("status %d, resp %v; want open-registration access", rec.Code, resp)
	}
	if created == nil || created.InvitedBy != auth.OpenRegistrationInviter || created.Email != "new@example.com" {
		t.Errorf("created invite = %+v", created)
	}
	if markedAccepted != "new@example.com" {
		t.Errorf("MarkAccepted called with %q, want new@example.com", markedAccepted)
	}
}

func TestInviteHandler_CheckAccess_SkipsMarkAcceptedWhenAlreadyAccepted(t *testing.T) {
	accepted := time.Now()
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(context.Context, string) (*domain.Invite, error) {
				return &domain.Invite{Email: "old@example.com", AcceptedAt: &accepted}, nil
			},
			MarkAcceptedFn: func(context.Context, string) (bool, error) {
				t.Error("MarkAccepted must not run for an already-accepted invite")
				return false, nil
			},
		},
	}
	req := httptest.NewRequest("GET", "/api/access", nil)
	req = req.WithContext(contextWithUserEmail("user-1", "old@example.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAdminStatsHandler_OpenRegistration(t *testing.T) {
	settings := memSettings{auth.OpenRegistrationSetting: "false"}
	h := &AdminStatsHandler{
		AppSettings: settings,
		Invites: &testutil.MockInviteRepo{
			CountByInviterFn: func(_ context.Context, invitedBy string) (int, error) {
				if invitedBy != auth.OpenRegistrationInviter {
					t.Errorf("counted invites by %q", invitedBy)
				}
				return 12, nil
			},
		},
	}

	call := func(fn http.HandlerFunc, method, body string) (int, map[string]any) {
		req := httptest.NewRequest(method, "/api/admin/settings/open-registration", strings.NewReader(body))
		rec := httptest.NewRecorder()
		fn(rec, req)
		var resp map[string]any
		json.NewDecoder(rec.Body).Decode(&resp)
		return rec.Code, resp
	}

	code, resp := call(h.GetOpenRegistration, "GET", "")
	if code != http.StatusOK || resp["openRegistration"] != false || resp["signups"] != float64(12) {
		t.Errorf("GET: %d %v", code, resp)
	}

	code, resp = call(h.SetOpenRegistration, "PATCH", `{"OpenRegistration":true}`)
	if code != http.StatusOK || resp["openRegistration"] != true || settings[auth.OpenRegistrationSetting] != "true" {
		t.Errorf("PATCH on: %d %v, stored %q", code, resp, settings[auth.OpenRegistrationSetting])
	}

	code, _ = call(h.SetOpenRegistration, "PATCH", `{"OpenRegistration":false}`)
	if code != http.StatusOK || settings[auth.OpenRegistrationSetting] != "false" {
		t.Errorf("PATCH off: %d, stored %q", code, settings[auth.OpenRegistrationSetting])
	}

	for _, body := range []string{`{}`, `{`, `{"OpenRegistration":"yes"}`} {
		if code, _ = call(h.SetOpenRegistration, "PATCH", body); code != http.StatusBadRequest {
			t.Errorf("PATCH %s: status %d, want 400", body, code)
		}
	}
	if settings[auth.OpenRegistrationSetting] != "false" {
		t.Error("invalid requests changed the setting")
	}
}
