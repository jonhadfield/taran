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

func TestDigestHandler_List_Success(t *testing.T) {
	h := &DigestHandler{
		Digests: &testutil.MockDigestRepo{
			ListFn: func(_ context.Context, _ string, _ domain.ListOptions) ([]domain.Digest, int, error) {
				return []domain.Digest{
					{ID: "d-1", Title: "Daily Digest"},
				}, 1, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/digests", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp ListResponse[domain.Digest]
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Data) != 1 {
		t.Errorf("data length = %d, want 1", len(resp.Data))
	}
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestDigestHandler_Get_Success(t *testing.T) {
	h := &DigestHandler{
		Digests: &testutil.MockDigestRepo{
			GetByIDFn: func(_ context.Context, _, id string) (*domain.Digest, error) {
				return &domain.Digest{ID: id, Title: "My Digest"}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/digests/d-1", nil)
	req.SetPathValue("id", "d-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp domain.Digest
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ID != "d-1" {
		t.Errorf("ID = %q, want %q", resp.ID, "d-1")
	}
}

func TestDigestHandler_Get_NotFound(t *testing.T) {
	h := &DigestHandler{
		Digests: &testutil.MockDigestRepo{
			GetByIDFn: func(_ context.Context, _, _ string) (*domain.Digest, error) {
				return nil, fmt.Errorf("not found")
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/digests/d-1", nil)
	req.SetPathValue("id", "d-1")
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
