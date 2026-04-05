package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type AdminStatsHandler struct {
	AdminStats  *database.AdminStatsRepo
	LLMProvider string
	LLMModel    string
	TokenUsage  database.TokenUsageRepository
	Preferences database.PreferenceRepository
	AppSettings *database.AppSettingRepo
	AuditLog    *database.AuditRepo
}

func (h *AdminStatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.AdminStats.GetStats(ctx)
	if err != nil {
		slog.Error("admin stats query failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to get admin stats")
		return
	}

	stats.LLMProvider = h.LLMProvider
	stats.LLMModel = h.LLMModel

	if h.AppSettings != nil {
		limit, _ := h.AppSettings.GetInt(ctx, "default_monthly_token_limit", 500000)
		stats.DefaultMonthlyTokenLimit = limit
	}

	if h.TokenUsage != nil {
		total, err := h.TokenUsage.GetGlobalMonthlyTotal(ctx)
		if err == nil {
			stats.MonthlyTokensUsed = total
		}
	}

	WriteJSON(w, http.StatusOK, stats)
}

func (h *AdminStatsHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.AdminStats.ListUsers(ctx)
	if err != nil {
		slog.Error("admin: failed to query users", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	defaultLimit := 500000
	if h.AppSettings != nil {
		if v, err := h.AppSettings.GetInt(ctx, "default_monthly_token_limit", 500000); err == nil {
			defaultLimit = v
		}
	}

	for i := range users {
		if users[i].MonthlyTokenLimit == 0 {
			users[i].MonthlyTokenLimit = defaultLimit
		}
	}

	WriteJSON(w, http.StatusOK, users)
}

func (h *AdminStatsHandler) SetDefaultTokenLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DefaultMonthlyTokenLimit int `json:"DefaultMonthlyTokenLimit"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DefaultMonthlyTokenLimit < 0 {
		WriteError(w, http.StatusBadRequest, "token limit cannot be negative")
		return
	}

	if err := h.AppSettings.Set(r.Context(), "default_monthly_token_limit", fmt.Sprintf("%d", req.DefaultMonthlyTokenLimit)); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update default token limit")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":                   "updated",
		"DefaultMonthlyTokenLimit": req.DefaultMonthlyTokenLimit,
	})
}

func (h *AdminStatsHandler) SetUserTokenLimit(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		WriteError(w, http.StatusBadRequest, "user id required")
		return
	}

	var req struct {
		MonthlyTokenLimit int `json:"MonthlyTokenLimit"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MonthlyTokenLimit < 0 {
		WriteError(w, http.StatusBadRequest, "token limit cannot be negative")
		return
	}

	pref, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get user preferences")
		return
	}

	pref.MonthlyTokenLimit = req.MonthlyTokenLimit
	if err := h.Preferences.Upsert(r.Context(), pref); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update token limit")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":            "updated",
		"UserID":            userID,
		"MonthlyTokenLimit": req.MonthlyTokenLimit,
	})
}

func (h *AdminStatsHandler) GetWaitlistEnabled(w http.ResponseWriter, r *http.Request) {
	enabled := false
	if h.AppSettings != nil {
		enabled, _ = h.AppSettings.GetBool(r.Context(), "waitlist_enabled", false)
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"waitlistEnabled": enabled})
}

func (h *AdminStatsHandler) SetWaitlistEnabled(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WaitlistEnabled bool `json:"WaitlistEnabled"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	value := "false"
	if req.WaitlistEnabled {
		value = "true"
	}

	if err := h.AppSettings.Set(r.Context(), "waitlist_enabled", value); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update waitlist setting")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":          "updated",
		"WaitlistEnabled": req.WaitlistEnabled,
	})
}

func (h *AdminStatsHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	if h.AuditLog == nil {
		WriteJSON(w, http.StatusOK, ListResponse[domain.AuditEntry]{Data: []domain.AuditEntry{}, Total: 0})
		return
	}

	entries, err := h.AuditLog.List(r.Context(), 100)
	if err != nil {
		slog.Error("failed to list audit log", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to list audit log")
		return
	}
	if entries == nil {
		entries = []domain.AuditEntry{}
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.AuditEntry]{Data: entries, Total: len(entries)})
}
