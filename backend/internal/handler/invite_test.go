package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func contextWithUserEmail(userID, email string) context.Context {
	ctx := testutil.ContextWithUserID(userID)
	return context.WithValue(ctx, auth.UserEmailKey, email)
}

func TestInviteHandler_Create_Success(t *testing.T) {
	var created *domain.Invite

	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.Invite, error) {
				return nil, nil // no existing invite
			},
			CreateFn: func(_ context.Context, invite *domain.Invite) error {
				created = invite
				return nil
			},
		},
		AdminEmails: []string{"admin@test.com"},
	}

	body := `{"email":"newuser@example.com"}`
	req := httptest.NewRequest("POST", "/api/invites", strings.NewReader(body))
	req = req.WithContext(contextWithUserEmail("user-1", "admin@test.com"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if created == nil {
		t.Fatal("expected Create to be called")
	}
	if created.Email != "newuser@example.com" {
		t.Errorf("Email = %q, want %q", created.Email, "newuser@example.com")
	}
	if created.InvitedBy != "admin@test.com" {
		t.Errorf("InvitedBy = %q, want %q", created.InvitedBy, "admin@test.com")
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestInviteHandler_Create_Duplicate(t *testing.T) {
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.Invite, error) {
				return &domain.Invite{ID: "existing"}, nil
			},
		},
		AdminEmails: []string{"admin@test.com"},
	}

	body := `{"email":"existing@example.com"}`
	req := httptest.NewRequest("POST", "/api/invites", strings.NewReader(body))
	req = req.WithContext(contextWithUserEmail("user-1", "admin@test.com"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestInviteHandler_Create_InvalidEmail(t *testing.T) {
	h := &InviteHandler{
		Invites:     &testutil.MockInviteRepo{},
		AdminEmails: []string{"admin@test.com"},
	}

	cases := []struct {
		name string
		body string
	}{
		{"empty email", `{"email":""}`},
		{"no at sign", `{"email":"notanemail"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/invites", strings.NewReader(tc.body))
			req = req.WithContext(contextWithUserEmail("user-1", "admin@test.com"))
			rec := httptest.NewRecorder()
			h.Create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestInviteHandler_List_Success(t *testing.T) {
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			ListFn: func(_ context.Context) ([]domain.Invite, error) {
				return []domain.Invite{
					{ID: "inv-1", Email: "user1@example.com"},
					{ID: "inv-2", Email: "user2@example.com"},
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/invites", nil)
	req = req.WithContext(contextWithUserEmail("user-1", "admin@test.com"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp ListResponse[domain.Invite]
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Errorf("data length = %d, want 2", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestInviteHandler_CheckAccess_Admin(t *testing.T) {
	h := &InviteHandler{
		Invites:     &testutil.MockInviteRepo{},
		AdminEmails: []string{"admin@test.com"},
	}

	req := httptest.NewRequest("GET", "/api/invites/check-access", nil)
	req = req.WithContext(contextWithUserEmail("user-1", "admin@test.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["hasAccess"] != true {
		t.Errorf("hasAccess = %v, want true", resp["hasAccess"])
	}
	if resp["reason"] != "admin" {
		t.Errorf("reason = %v, want %q", resp["reason"], "admin")
	}
}

func TestInviteHandler_CheckAccess_Invited(t *testing.T) {
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.Invite, error) {
				return &domain.Invite{ID: "inv-1", Email: "invited@example.com"}, nil
			},
		},
		AdminEmails: []string{"admin@test.com"},
	}

	req := httptest.NewRequest("GET", "/api/invites/check-access", nil)
	req = req.WithContext(contextWithUserEmail("user-2", "invited@example.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["hasAccess"] != true {
		t.Errorf("hasAccess = %v, want true", resp["hasAccess"])
	}
	if resp["reason"] != "invited" {
		t.Errorf("reason = %v, want %q", resp["reason"], "invited")
	}
}

func TestInviteHandler_CheckAccess_NotInvited(t *testing.T) {
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(_ context.Context, _ string) (*domain.Invite, error) {
				return nil, nil
			},
		},
		AdminEmails: []string{"admin@test.com"},
	}

	req := httptest.NewRequest("GET", "/api/invites/check-access", nil)
	req = req.WithContext(contextWithUserEmail("user-3", "random@example.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["hasAccess"] != false {
		t.Errorf("hasAccess = %v, want false", resp["hasAccess"])
	}
	if resp["reason"] != "not_invited" {
		t.Errorf("reason = %v, want %q", resp["reason"], "not_invited")
	}
}
