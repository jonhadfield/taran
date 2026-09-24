package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
)

// fakeAdminStats records the page it was asked for.
type fakeAdminStats struct {
	total       int
	gotLimit    int
	gotOffset   int
	returnUsers []domain.AdminUser
}

func (f *fakeAdminStats) GetStats(context.Context) (*domain.AdminStats, error) {
	return &domain.AdminStats{}, nil
}

func (f *fakeAdminStats) ListUsers(_ context.Context, limit, offset int) ([]domain.AdminUser, int, error) {
	f.gotLimit, f.gotOffset = limit, offset
	return f.returnUsers, f.total, nil
}

func listUsers(t *testing.T, stats *fakeAdminStats, query string) (int, map[string]any) {
	t.Helper()
	h := &AdminStatsHandler{AdminStats: stats}
	req := httptest.NewRequest("GET", "/api/admin/users"+query, nil)
	rec := httptest.NewRecorder()
	h.ListUsers(rec, req)
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	return rec.Code, resp
}

func TestAdminListUsers_DefaultPage(t *testing.T) {
	stats := &fakeAdminStats{total: 73, returnUsers: []domain.AdminUser{{ID: "u1"}}}
	code, resp := listUsers(t, stats, "")

	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if stats.gotLimit != defaultUserPageSize || stats.gotOffset != 0 {
		t.Errorf("asked for limit %d offset %d, want %d and 0", stats.gotLimit, stats.gotOffset, defaultUserPageSize)
	}
	if resp["total"] != float64(73) {
		t.Errorf("total = %v, want 73", resp["total"])
	}
	if data, ok := resp["data"].([]any); !ok || len(data) != 1 {
		t.Errorf("data = %v, want one user", resp["data"])
	}
}

func TestAdminListUsers_PagingParams(t *testing.T) {
	tests := []struct {
		query         string
		limit, offset int
	}{
		{"?limit=10&offset=30", 10, 30},
		{"?limit=1000", maxUserPageSize, 0},               // clamped
		{"?limit=0", 1, 0},                                // clamped
		{"?limit=-5&offset=-5", 1, 0},                     // negatives rejected
		{"?limit=abc&offset=xyz", defaultUserPageSize, 0}, // non-numeric falls back
	}
	for _, tt := range tests {
		stats := &fakeAdminStats{}
		if code, _ := listUsers(t, stats, tt.query); code != http.StatusOK {
			t.Fatalf("%s: status = %d", tt.query, code)
		}
		if stats.gotLimit != tt.limit || stats.gotOffset != tt.offset {
			t.Errorf("%s: limit/offset = %d/%d, want %d/%d", tt.query, stats.gotLimit, stats.gotOffset, tt.limit, tt.offset)
		}
	}
}

func TestAdminListUsers_EmptyPageIsAnArray(t *testing.T) {
	stats := &fakeAdminStats{total: 0, returnUsers: []domain.AdminUser{}}
	_, resp := listUsers(t, stats, "?offset=999")
	data, ok := resp["data"].([]any)
	if !ok || len(data) != 0 {
		t.Errorf("data = %v, want an empty array", resp["data"])
	}
}
