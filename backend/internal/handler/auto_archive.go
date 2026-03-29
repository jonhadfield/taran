package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type AutoArchiveHandler struct {
	Rules database.AutoArchiveRuleRepository
}

func (h *AutoArchiveHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	rules, err := h.Rules.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}
	if rules == nil {
		rules = []domain.AutoArchiveRule{}
	}
	WriteJSON(w, http.StatusOK, rules)
}

type upsertAutoArchiveRequest struct {
	RuleType         string `json:"RuleType"`
	RuleValue        string `json:"RuleValue"`
	ArchiveAfterDays int    `json:"ArchiveAfterDays"`
	IsActive         *bool  `json:"IsActive"`
}

func (h *AutoArchiveHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req upsertAutoArchiveRequest
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RuleType != "category" && req.RuleType != "sender" {
		WriteError(w, http.StatusBadRequest, "rule_type must be 'category' or 'sender'")
		return
	}
	if req.RuleValue == "" {
		WriteError(w, http.StatusBadRequest, "rule_value is required")
		return
	}
	if req.RuleType == "category" {
		validCats := map[string]bool{"newsletter": true, "personal": true, "transactional": true, "marketing": true, "notification": true, "other": true}
		if !validCats[req.RuleValue] {
			WriteError(w, http.StatusBadRequest, "invalid category value")
			return
		}
	}
	if req.ArchiveAfterDays < 1 || req.ArchiveAfterDays > 365 {
		WriteError(w, http.StatusBadRequest, "archive_after_days must be between 1 and 365")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &domain.AutoArchiveRule{
		ID:               uuid.New().String(),
		UserID:           userID,
		RuleType:         req.RuleType,
		RuleValue:        req.RuleValue,
		ArchiveAfterDays: req.ArchiveAfterDays,
		IsActive:         isActive,
	}

	if err := h.Rules.Upsert(r.Context(), rule); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to save rule")
		return
	}

	// Return updated list
	rules, err := h.Rules.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}
	WriteJSON(w, http.StatusOK, rules)
}

func (h *AutoArchiveHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing rule id")
		return
	}

	if err := h.Rules.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete rule")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
