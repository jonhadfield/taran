package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

const (
	maxAnalysisRules      = 20
	maxAnalysisRuleLength = 300
)

type AnalysisRuleHandler struct {
	AnalysisRules database.AnalysisRuleRepository
}

func (h *AnalysisRuleHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	h.writeList(w, r, userID)
}

func (h *AnalysisRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		Rule string `json:"Rule"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	text, msg := validateAnalysisRule(req.Rule)
	if msg != "" {
		WriteError(w, http.StatusBadRequest, msg)
		return
	}

	count, err := h.AnalysisRules.CountByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check analysis rule limit")
		return
	}
	if count >= maxAnalysisRules {
		WriteError(w, http.StatusBadRequest, "maximum 20 analysis rules allowed")
		return
	}

	rule := &domain.AnalysisRule{
		ID:       uuid.New().String(),
		UserID:   userID,
		Rule:     text,
		IsActive: true,
	}
	if err := h.AnalysisRules.Create(r.Context(), rule); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create analysis rule")
		return
	}

	h.writeList(w, r, userID)
}

func (h *AnalysisRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing analysis rule id")
		return
	}

	var req struct {
		Rule     *string `json:"Rule"`
		IsActive *bool   `json:"IsActive"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Rule == nil && req.IsActive == nil {
		WriteError(w, http.StatusBadRequest, "nothing to update")
		return
	}
	if req.Rule != nil {
		text, msg := validateAnalysisRule(*req.Rule)
		if msg != "" {
			WriteError(w, http.StatusBadRequest, msg)
			return
		}
		req.Rule = &text
	}

	if err := h.AnalysisRules.Update(r.Context(), userID, id, req.Rule, req.IsActive); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			WriteError(w, http.StatusNotFound, "analysis rule not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to update analysis rule")
		return
	}

	h.writeList(w, r, userID)
}

func (h *AnalysisRuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing analysis rule id")
		return
	}

	if err := h.AnalysisRules.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete analysis rule")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalysisRuleHandler) writeList(w http.ResponseWriter, r *http.Request, userID string) {
	rules, err := h.AnalysisRules.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list analysis rules")
		return
	}
	if rules == nil {
		rules = []domain.AnalysisRule{}
	}
	WriteJSON(w, http.StatusOK, rules)
}

// validateAnalysisRule trims the rule and returns it, or a non-empty error
// message if it is unacceptable.
func validateAnalysisRule(rule string) (string, string) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return "", "rule is required"
	}
	if utf8.RuneCountInString(rule) > maxAnalysisRuleLength {
		return "", "rule must be 300 characters or fewer"
	}
	return rule, ""
}
