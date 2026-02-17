package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type SenderHandler struct {
	Emails      database.EmailRepository
	SenderPrefs database.SenderPreferenceRepository
	Feedback    database.FeedbackRepository
}

func (h *SenderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	// Get all emails for the user to aggregate senders
	emails, _, err := h.Emails.List(r.Context(), userID, domain.ListOptions{Limit: 10000})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list emails")
		return
	}

	// Aggregate senders
	type senderAgg struct {
		FromName string
		Count    int
	}
	senderMap := make(map[string]*senderAgg)
	for _, e := range emails {
		agg, ok := senderMap[e.FromAddress]
		if !ok {
			senderMap[e.FromAddress] = &senderAgg{FromName: e.FromName, Count: 1}
		} else {
			agg.Count++
			if agg.FromName == "" && e.FromName != "" {
				agg.FromName = e.FromName
			}
		}
	}

	// Get sender preferences
	prefs, err := h.SenderPrefs.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list sender preferences")
		return
	}
	prefMap := make(map[string]string)
	for _, p := range prefs {
		prefMap[p.FromAddress] = p.Status
	}

	// Build response
	var senders []domain.SenderInfo
	for addr, agg := range senderMap {
		status := "normal"
		if s, ok := prefMap[addr]; ok {
			status = s
		}
		senders = append(senders, domain.SenderInfo{
			FromAddress: addr,
			FromName:    agg.FromName,
			EmailCount:  agg.Count,
			Status:      status,
		})
	}

	WriteJSON(w, http.StatusOK, senders)
}

func (h *SenderHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var body struct {
		FromAddress string `json:"FromAddress"`
		Status      string `json:"Status"`
	}
	if err := LimitedJSONDecoder(r).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.FromAddress == "" {
		WriteError(w, http.StatusBadRequest, "FromAddress is required")
		return
	}

	validStatuses := map[string]bool{"normal": true, "muted": true, "blocked": true, "favorite": true}
	if !validStatuses[body.Status] {
		WriteError(w, http.StatusBadRequest, "Status must be one of: normal, muted, blocked, favorite")
		return
	}

	now := time.Now()
	pref := &domain.SenderPreference{
		ID:          uuid.New().String(),
		UserID:      userID,
		FromAddress: body.FromAddress,
		Status:      body.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.SenderPrefs.Upsert(r.Context(), pref); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update sender preference")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

type senderSuggestion struct {
	FromAddress    string `json:"FromAddress"`
	FromName       string `json:"FromName"`
	NotUsefulCount int    `json:"NotUsefulCount"`
	TotalCount     int    `json:"TotalCount"`
}

func (h *SenderHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	if h.Feedback == nil {
		WriteJSON(w, http.StatusOK, []senderSuggestion{})
		return
	}

	stats, err := h.Feedback.GetSenderStats(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get feedback stats")
		return
	}

	// Get current sender preferences to exclude already muted/blocked
	prefs, err := h.SenderPrefs.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list sender preferences")
		return
	}
	excludeMap := make(map[string]bool)
	for _, p := range prefs {
		if p.Status == "muted" || p.Status == "blocked" {
			excludeMap[p.FromAddress] = true
		}
	}

	// Get sender names from email data
	emails, _, err := h.Emails.List(r.Context(), userID, domain.ListOptions{Limit: 10000})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list emails")
		return
	}
	nameMap := make(map[string]string)
	for _, e := range emails {
		if _, ok := nameMap[e.FromAddress]; !ok && e.FromName != "" {
			nameMap[e.FromAddress] = e.FromName
		}
	}

	var suggestions []senderSuggestion
	for _, stat := range stats {
		if excludeMap[stat.FromAddress] {
			continue
		}
		total := stat.UsefulCount + stat.NotUsefulCount
		if total < 3 {
			continue
		}
		notUsefulPct := float64(stat.NotUsefulCount) / float64(total)
		if notUsefulPct <= 0.6 {
			continue
		}
		suggestions = append(suggestions, senderSuggestion{
			FromAddress:    stat.FromAddress,
			FromName:       nameMap[stat.FromAddress],
			NotUsefulCount: stat.NotUsefulCount,
			TotalCount:     total,
		})
	}

	if suggestions == nil {
		suggestions = []senderSuggestion{}
	}

	WriteJSON(w, http.StatusOK, suggestions)
}
