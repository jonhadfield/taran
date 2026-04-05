package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
)

type UsageHandler struct {
	TokenUsage  database.TokenUsageRepository
	Preferences database.PreferenceRepository
}

func (h *UsageHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	stats, err := h.TokenUsage.GetUsageStats(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load usage stats")
		return
	}

	// Fill in limits from preferences
	pref, err := h.Preferences.Get(r.Context(), userID)
	if err == nil {
		stats.MonthlyTokenLimit = pref.MonthlyTokenLimit
		stats.DailyTokenLimit = pref.DailyTokenLimit
	}

	// Include 30-day daily breakdown
	history, err := h.TokenUsage.GetDailyBreakdown(r.Context(), userID, 30)
	if err == nil {
		stats.DailyHistory = history
	}

	WriteJSON(w, http.StatusOK, stats)
}
