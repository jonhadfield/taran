package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
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
	DigestEmail bool `json:"DigestEmail"`
}

func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req updatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	pref := &domain.UserPreference{
		UserID:      userID,
		DigestEmail: req.DigestEmail,
	}

	if err := h.Preferences.Upsert(r.Context(), pref); err != nil {
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
