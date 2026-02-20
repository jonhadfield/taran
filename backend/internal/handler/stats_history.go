package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type StatsHistoryHandler struct {
	Emails database.EmailRepository
}

func (h *StatsHistoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	weeks := intParam(r, "weeks", 8)
	if weeks < 1 {
		weeks = 1
	}
	if weeks > 52 {
		weeks = 52
	}

	counts, err := h.Emails.CountByWeek(r.Context(), userID, weeks)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	if counts == nil {
		counts = []domain.WeekCount{}
	}

	WriteJSON(w, http.StatusOK, counts)
}
