package handler

import (
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type WeeklySummaryHandler struct {
	Summaries *database.WeeklySummaryRepo
	Runner    *digest.WeeklySummaryRunner
}

func (h *WeeklySummaryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	limit := clampInt(intParam(r, "limit", 50), 1, 200)
	offset := clampInt(intParam(r, "offset", 0), 0, 100000)

	summaries, total, err := h.Summaries.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list weekly summaries")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.WeeklySummary]{Data: summaries, Total: total})
}

func (h *WeeklySummaryHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	s, err := h.Summaries.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "weekly summary not found")
		return
	}

	WriteJSON(w, http.StatusOK, s)
}

func (h *WeeklySummaryHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	now := time.Now().UTC()
	periodEnd := now
	periodStart := now.Add(-7 * 24 * time.Hour)

	summary, err := h.Runner.GenerateForUser(r.Context(), userID, periodStart, periodEnd)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate weekly summary")
		return
	}
	if summary == nil {
		WriteError(w, http.StatusUnprocessableEntity, "no email activity in the last 7 days")
		return
	}

	if err := h.Summaries.Create(r.Context(), summary); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to store weekly summary")
		return
	}

	WriteJSON(w, http.StatusOK, summary)
}
