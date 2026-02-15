package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type EmailHandler struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
}

type EmailResponse struct {
	domain.Email
	Extraction *domain.Extraction `json:"extraction,omitempty"`
}

func (h *EmailHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	opts := domain.ListOptions{
		Limit:  intParam(r, "limit", 50),
		Offset: intParam(r, "offset", 0),
	}

	if v := r.URL.Query().Get("is_read"); v != "" {
		b := v == "true"
		opts.IsRead = &b
	}
	if v := r.URL.Query().Get("is_starred"); v != "" {
		b := v == "true"
		opts.IsStarred = &b
	}
	if v := r.URL.Query().Get("is_archived"); v != "" {
		b := v == "true"
		opts.IsArchived = &b
	}
	if v := r.URL.Query().Get("search"); v != "" {
		opts.Search = &v
	}

	emails, total, err := h.Emails.List(r.Context(), userID, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list emails")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.Email]{Data: emails, Total: total})
}

func (h *EmailHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	email, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	resp := EmailResponse{Email: *email}
	extraction, err := h.Extractions.GetByEmailID(r.Context(), id)
	if err == nil {
		resp.Extraction = extraction
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *EmailHandler) UpdateState(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var state domain.EmailState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.Emails.UpdateState(r.Context(), userID, id, state); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update email")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func intParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
