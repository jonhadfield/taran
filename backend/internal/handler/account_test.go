package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

const testDomain = "mailbrief.io"

func TestAccountHandler_List_Success(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			ListByUserFn: func(_ context.Context, _ string) ([]domain.EmailAccount, error) {
				return []domain.EmailAccount{
					{ID: "acct-1", EmailAddress: "user@test.com"},
					{ID: "acct-2", EmailAddress: "user2@test.com"},
				}, nil
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp ListResponse[domain.EmailAccount]
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 2 {
		t.Errorf("data length = %d, want 2", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestAccountHandler_List_Error(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			ListByUserFn: func(_ context.Context, _ string) ([]domain.EmailAccount, error) {
				return nil, fmt.Errorf("db error")
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAccountHandler_Create_Success(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return nil, nil
			},
			CreateFn: func(_ context.Context, _ *domain.EmailAccount) error {
				return nil
			},
		},
		EmailDomain: testDomain,
	}

	body := `{"Username":"jon"}`
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var account domain.EmailAccount
	json.NewDecoder(rec.Body).Decode(&account)
	if account.EmailAddress != "jon@mailbrief.io" {
		t.Errorf("email = %q, want %q", account.EmailAddress, "jon@mailbrief.io")
	}
	if account.DisplayName != "jon" {
		t.Errorf("display name = %q, want %q", account.DisplayName, "jon")
	}
	if !account.IsActive {
		t.Error("expected IsActive to be true")
	}
	if account.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestAccountHandler_Create_MissingUsername(t *testing.T) {
	h := &AccountHandler{
		Accounts:    &testutil.MockAccountRepo{},
		EmailDomain: testDomain,
	}

	body := `{}`
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAccountHandler_Create_InvalidUsername(t *testing.T) {
	h := &AccountHandler{
		Accounts:    &testutil.MockAccountRepo{},
		EmailDomain: testDomain,
	}

	cases := []struct {
		name     string
		username string
	}{
		{"too short", "ab"},
		{"leading hyphen", "-abc"},
		{"trailing hyphen", "abc-"},
		{"special chars", "abc@def"},
		{"spaces", "ab cd"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"Username":%q}`, tc.username)
			req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
			req = req.WithContext(testutil.ContextWithUserID("user-1"))
			rec := httptest.NewRecorder()
			h.Create(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d for username %q", rec.Code, http.StatusBadRequest, tc.username)
			}
		})
	}
}

func TestAccountHandler_Create_DuplicateUsername(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "existing"}, nil
			},
		},
		EmailDomain: testDomain,
	}

	body := `{"Username":"jon"}`
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAccountHandler_Create_Error(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return nil, nil
			},
			CreateFn: func(_ context.Context, _ *domain.EmailAccount) error {
				return fmt.Errorf("db error")
			},
		},
		EmailDomain: testDomain,
	}

	body := `{"Username":"jon"}`
	req := httptest.NewRequest("POST", "/api/accounts", bytes.NewBufferString(body))
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestAccountHandler_CheckUsername_Available(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return nil, nil
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts/check-username?username=jon", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.CheckUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["available"] != true {
		t.Errorf("available = %v, want true", resp["available"])
	}
}

func TestAccountHandler_CheckUsername_Taken(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "existing"}, nil
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts/check-username?username=jon", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.CheckUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["available"] != false {
		t.Errorf("available = %v, want false", resp["available"])
	}
}

func TestAccountHandler_CheckUsername_Invalid(t *testing.T) {
	h := &AccountHandler{
		Accounts:    &testutil.MockAccountRepo{},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts/check-username?username=ab", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.CheckUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["available"] != false {
		t.Errorf("available = %v, want false", resp["available"])
	}
	if resp["reason"] == nil {
		t.Error("expected reason to be set for invalid username")
	}
}

func TestAccountHandler_CheckUsername_Missing(t *testing.T) {
	h := &AccountHandler{
		Accounts:    &testutil.MockAccountRepo{},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("GET", "/api/accounts/check-username", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.CheckUsername(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAccountHandler_Delete_Success(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			DeleteFn: func(_ context.Context, _, _ string) error {
				return nil
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("DELETE", "/api/accounts/acct-1", nil)
	req.SetPathValue("id", "acct-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAccountHandler_Delete_Error(t *testing.T) {
	h := &AccountHandler{
		Accounts: &testutil.MockAccountRepo{
			DeleteFn: func(_ context.Context, _, _ string) error {
				return fmt.Errorf("account not found")
			},
		},
		EmailDomain: testDomain,
	}

	req := httptest.NewRequest("DELETE", "/api/accounts/acct-1", nil)
	req.SetPathValue("id", "acct-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
