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

	senders, err := h.Emails.ListSenders(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list senders")
		return
	}

	// Merge sender preference status and category overrides
	prefs, err := h.SenderPrefs.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list sender preferences")
		return
	}
	type prefInfo struct {
		Status   string
		Category string
	}
	prefMap := make(map[string]prefInfo)
	for _, p := range prefs {
		prefMap[p.FromAddress] = prefInfo{Status: p.Status, Category: p.Category}
	}

	for i := range senders {
		if p, ok := prefMap[senders[i].FromAddress]; ok {
			senders[i].Status = p.Status
			if p.Category != "" {
				senders[i].Category = p.Category
			} else {
				senders[i].Category = senders[i].AutoCategory
			}
		} else {
			senders[i].Status = "normal"
			senders[i].Category = senders[i].AutoCategory
		}
	}

	WriteJSON(w, http.StatusOK, senders)
}

func (h *SenderHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var body struct {
		FromAddress string `json:"FromAddress"`
		Status      string `json:"Status"`
		Category    string `json:"Category"`
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

	validCategories := map[string]bool{"": true, "newsletter": true, "personal": true, "transactional": true, "marketing": true, "notification": true, "other": true}
	if !validCategories[body.Category] {
		WriteError(w, http.StatusBadRequest, "Category must be one of: newsletter, personal, transactional, marketing, notification, other (or empty for auto)")
		return
	}

	now := time.Now()
	pref := &domain.SenderPreference{
		ID:          uuid.New().String(),
		UserID:      userID,
		FromAddress: body.FromAddress,
		Status:      body.Status,
		Category:    body.Category,
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

	// Get sender names via SQL aggregation
	senderInfos, err := h.Emails.ListSenders(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list senders")
		return
	}
	nameMap := make(map[string]string)
	for _, s := range senderInfos {
		if s.FromName != "" {
			nameMap[s.FromAddress] = s.FromName
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
