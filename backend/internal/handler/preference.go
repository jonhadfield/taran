package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
)

type PreferenceHandler struct {
	Preferences database.PreferenceRepository
}

func (h *PreferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	pref, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	WriteJSON(w, http.StatusOK, pref)
}

type updatePreferenceRequest struct {
	DigestEmail     *bool   `json:"DigestEmail"`
	DigestFrequency *string `json:"DigestFrequency"`
	DigestHour      *int    `json:"DigestHour"`
	DigestTimezone  *string `json:"DigestTimezone"`
}

func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req updatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate frequency
	if req.DigestFrequency != nil {
		if *req.DigestFrequency != "daily" && *req.DigestFrequency != "weekly" {
			WriteError(w, http.StatusBadRequest, "frequency must be 'daily' or 'weekly'")
			return
		}
	}

	// Validate hour
	if req.DigestHour != nil {
		if *req.DigestHour < 0 || *req.DigestHour > 23 {
			WriteError(w, http.StatusBadRequest, "hour must be between 0 and 23")
			return
		}
	}

	// Validate timezone
	if req.DigestTimezone != nil {
		if _, err := time.LoadLocation(*req.DigestTimezone); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid timezone")
			return
		}
	}

	// Load existing preference and merge
	existing, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	if req.DigestEmail != nil {
		existing.DigestEmail = *req.DigestEmail
	}
	if req.DigestFrequency != nil {
		existing.DigestFrequency = *req.DigestFrequency
	}
	if req.DigestHour != nil {
		existing.DigestHour = *req.DigestHour
	}
	if req.DigestTimezone != nil {
		existing.DigestTimezone = *req.DigestTimezone
	}

	if err := h.Preferences.Upsert(r.Context(), existing); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	updated, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}
