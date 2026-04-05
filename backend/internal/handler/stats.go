package handler

import (
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type StatsHandler struct {
	Emails database.EmailRepository
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	d := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, t.Location())
}

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	ctx := r.Context()
	now := time.Now()

	weekStart := startOfWeek(now)
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	thisWeek, err := h.Emails.CountByPeriod(ctx, userID, weekStart, now)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to count emails")
		return
	}

	lastWeek, err := h.Emails.CountByPeriod(ctx, userID, lastWeekStart, weekStart)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to count emails")
		return
	}

	processed := domain.EmailStatusProcessed
	total, err := h.Emails.CountByFilter(ctx, userID, &processed, nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to count total emails")
		return
	}

	topSenders, err := h.Emails.TopSenders(ctx, userID, weekStart, now, 5)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get top senders")
		return
	}

	WriteJSON(w, http.StatusOK, domain.UserStats{
		EmailsThisWeek: thisWeek,
		EmailsLastWeek: lastWeek,
		TotalEmails:    total,
		TopSenders:     topSenders,
	})
}
