package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

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
	}

	req := httptest.NewRequest("GET", "/api/accounts", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
