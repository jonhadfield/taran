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
