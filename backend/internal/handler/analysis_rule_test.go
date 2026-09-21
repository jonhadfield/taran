package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

// fakeRuleRepo is an in-memory AnalysisRuleRepository that enforces user scoping
// the same way the SQL implementation does.
type fakeRuleRepo struct {
	rules []domain.AnalysisRule
}

func (f *fakeRuleRepo) Create(_ context.Context, r *domain.AnalysisRule) error {
	f.rules = append(f.rules, *r)
	return nil
}

func (f *fakeRuleRepo) ListByUser(_ context.Context, userID string) ([]domain.AnalysisRule, error) {
	var out []domain.AnalysisRule
	for _, r := range f.rules {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRuleRepo) ListActiveRules(_ context.Context, userID string) ([]string, error) {
	var out []string
	for _, r := range f.rules {
		if r.UserID == userID && r.IsActive {
			out = append(out, r.Rule)
		}
	}
	return out, nil
}

func (f *fakeRuleRepo) Update(_ context.Context, userID, id string, rule *string, isActive *bool) error {
	for i, r := range f.rules {
		if r.ID == id && r.UserID == userID {
			if rule != nil {
				f.rules[i].Rule = *rule
			}
			if isActive != nil {
				f.rules[i].IsActive = *isActive
			}
			return nil
		}
	}
	return database.ErrNotFound
}

func (f *fakeRuleRepo) Delete(_ context.Context, userID, id string) error {
	for i, r := range f.rules {
		if r.ID == id && r.UserID == userID {
			f.rules = append(f.rules[:i], f.rules[i+1:]...)
			return nil
		}
	}
	return nil
}

func (f *fakeRuleRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	rules, _ := f.ListByUser(ctx, userID)
	return len(rules), nil
}

func doRuleRequest(t *testing.T, fn http.HandlerFunc, method, path, id, userID, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(testutil.ContextWithUserID(userID))
	if id != "" {
		req.SetPathValue("id", id)
	}
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}

func decodeRules(t *testing.T, rec *httptest.ResponseRecorder) []domain.AnalysisRule {
	t.Helper()
	var rules []domain.AnalysisRule
	if err := json.NewDecoder(rec.Body).Decode(&rules); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return rules
}

func TestAnalysisRuleHandler_ListEmptyReturnsArray(t *testing.T) {
	h := &AnalysisRuleHandler{AnalysisRules: &fakeRuleRepo{}}
	rec := doRuleRequest(t, h.List, "GET", "/api/analysis-rules", "", "user-1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func TestAnalysisRuleHandler_CreateTrimsAndActivates(t *testing.T) {
	h := &AnalysisRuleHandler{AnalysisRules: &fakeRuleRepo{}}
	rec := doRuleRequest(t, h.Create, "POST", "/api/analysis-rules", "", "user-1",
		`{"Rule":"  More detail on ISA, Bitcoin and Ethereum  "}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	rules := decodeRules(t, rec)
	if len(rules) != 1 || rules[0].Rule != "More detail on ISA, Bitcoin and Ethereum" || !rules[0].IsActive || rules[0].ID == "" {
		t.Errorf("unexpected rules: %+v", rules)
	}
}

func TestAnalysisRuleHandler_CreateValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"empty", `{"Rule":""}`},
		{"whitespace", `{"Rule":"   "}`},
		{"too long", `{"Rule":"` + strings.Repeat("a", 301) + `"}`},
		{"bad json", `{`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRuleRepo{}
			h := &AnalysisRuleHandler{AnalysisRules: repo}
			rec := doRuleRequest(t, h.Create, "POST", "/api/analysis-rules", "", "user-1", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
			if len(repo.rules) != 0 {
				t.Error("rule was created despite invalid input")
			}
		})
	}
}

func TestAnalysisRuleHandler_Create300RunesAllowed(t *testing.T) {
	h := &AnalysisRuleHandler{AnalysisRules: &fakeRuleRepo{}}
	// Multi-byte characters count as one each.
	rec := doRuleRequest(t, h.Create, "POST", "/api/analysis-rules", "", "user-1",
		`{"Rule":"`+strings.Repeat("£", 300)+`"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAnalysisRuleHandler_CreateEnforcesLimit(t *testing.T) {
	repo := &fakeRuleRepo{}
	for i := 0; i < maxAnalysisRules; i++ {
		repo.rules = append(repo.rules, domain.AnalysisRule{ID: string(rune('a' + i)), UserID: "user-1", Rule: "r"})
	}
	// Another user's rules don't count towards the limit.
	repo.rules = append(repo.rules, domain.AnalysisRule{ID: "other", UserID: "user-2", Rule: "r"})
	h := &AnalysisRuleHandler{AnalysisRules: repo}

	rec := doRuleRequest(t, h.Create, "POST", "/api/analysis-rules", "", "user-1", `{"Rule":"one more"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "maximum 20") {
		t.Errorf("status = %d, body = %s; want 400 limit error", rec.Code, rec.Body)
	}

	rec = doRuleRequest(t, h.Create, "POST", "/api/analysis-rules", "", "user-2", `{"Rule":"fine"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("user-2 status = %d, want 200", rec.Code)
	}
}

func TestAnalysisRuleHandler_Update(t *testing.T) {
	repo := &fakeRuleRepo{rules: []domain.AnalysisRule{{ID: "r1", UserID: "user-1", Rule: "old", IsActive: true}}}
	h := &AnalysisRuleHandler{AnalysisRules: repo}

	rec := doRuleRequest(t, h.Update, "PATCH", "/api/analysis-rules/r1", "r1", "user-1", `{"IsActive":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if rules := decodeRules(t, rec); rules[0].IsActive || rules[0].Rule != "old" {
		t.Errorf("toggle changed the wrong fields: %+v", rules[0])
	}

	rec = doRuleRequest(t, h.Update, "PATCH", "/api/analysis-rules/r1", "r1", "user-1", `{"Rule":" new "}`)
	if rules := decodeRules(t, rec); rules[0].Rule != "new" || rules[0].IsActive {
		t.Errorf("edit changed the wrong fields: %+v", rules[0])
	}

	for name, body := range map[string]string{"empty rule": `{"Rule":""}`, "no fields": `{}`} {
		rec = doRuleRequest(t, h.Update, "PATCH", "/api/analysis-rules/r1", "r1", "user-1", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
}

func TestAnalysisRuleHandler_UpdateOtherUsersRuleIsNotFound(t *testing.T) {
	repo := &fakeRuleRepo{rules: []domain.AnalysisRule{{ID: "r1", UserID: "user-1", Rule: "mine", IsActive: true}}}
	h := &AnalysisRuleHandler{AnalysisRules: repo}

	rec := doRuleRequest(t, h.Update, "PATCH", "/api/analysis-rules/r1", "r1", "user-2", `{"Rule":"hijacked"}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if repo.rules[0].Rule != "mine" {
		t.Error("another user's rule was modified")
	}
}

func TestAnalysisRuleHandler_DeleteIsScoped(t *testing.T) {
	repo := &fakeRuleRepo{rules: []domain.AnalysisRule{{ID: "r1", UserID: "user-1", Rule: "mine"}}}
	h := &AnalysisRuleHandler{AnalysisRules: repo}

	rec := doRuleRequest(t, h.Delete, "DELETE", "/api/analysis-rules/r1", "r1", "user-2", "")
	if rec.Code != http.StatusNoContent || len(repo.rules) != 1 {
		t.Errorf("another user deleted the rule (status %d, %d left)", rec.Code, len(repo.rules))
	}

	rec = doRuleRequest(t, h.Delete, "DELETE", "/api/analysis-rules/r1", "r1", "user-1", "")
	if rec.Code != http.StatusNoContent || len(repo.rules) != 0 {
		t.Errorf("owner could not delete the rule (status %d, %d left)", rec.Code, len(repo.rules))
	}
}
