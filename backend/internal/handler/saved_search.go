package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type SavedSearchHandler struct {
	SavedSearches database.SavedSearchRepository
}

func (h *SavedSearchHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	searches, err := h.SavedSearches.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list saved searches")
		return
	}
	if searches == nil {
		searches = []domain.SavedSearch{}
	}
	WriteJSON(w, http.StatusOK, searches)
}

func (h *SavedSearchHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		Name    string                 `json:"Name"`
		Filters map[string]interface{} `json:"Filters"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Name) > 100 {
		WriteError(w, http.StatusBadRequest, "name must be 100 characters or fewer")
		return
	}

	count, err := h.SavedSearches.CountByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check saved search limit")
		return
	}
	if count >= 20 {
		WriteError(w, http.StatusBadRequest, "maximum 20 saved searches allowed")
		return
	}

	s := &domain.SavedSearch{
		ID:      uuid.New().String(),
		UserID:  userID,
		Name:    req.Name,
		Filters: req.Filters,
	}
	if err := h.SavedSearches.Create(r.Context(), s); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create saved search")
		return
	}

	// Return updated list
	searches, err := h.SavedSearches.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list saved searches")
		return
	}
	WriteJSON(w, http.StatusOK, searches)
}

func (h *SavedSearchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing saved search id")
		return
	}

	if err := h.SavedSearches.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete saved search")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
