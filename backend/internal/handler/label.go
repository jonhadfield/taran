package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type LabelHandler struct {
	Labels database.LabelRepository
}

func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	labels, err := h.Labels.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list labels")
		return
	}
	if labels == nil {
		labels = []domain.Label{}
	}
	WriteJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		Name  string `json:"Name"`
		Color string `json:"Color"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Name) > 50 {
		WriteError(w, http.StatusBadRequest, "name must be 50 characters or fewer")
		return
	}

	label := &domain.Label{
		ID:     uuid.New().String(),
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
	}
	if err := h.Labels.Create(r.Context(), label); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create label")
		return
	}

	// Return updated list
	labels, err := h.Labels.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list labels")
		return
	}
	WriteJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing label id")
		return
	}

	var req struct {
		Name  string `json:"Name"`
		Color string `json:"Color"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Name) > 50 {
		WriteError(w, http.StatusBadRequest, "name must be 50 characters or fewer")
		return
	}

	label := &domain.Label{
		ID:    id,
		Name:  req.Name,
		Color: req.Color,
	}
	if err := h.Labels.Update(r.Context(), userID, label); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update label")
		return
	}

	// Return updated list
	labels, err := h.Labels.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list labels")
		return
	}
	WriteJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing label id")
		return
	}

	if err := h.Labels.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete label")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) AddToEmail(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	emailID := r.PathValue("id")
	labelID := r.PathValue("labelId")
	if emailID == "" || labelID == "" {
		WriteError(w, http.StatusBadRequest, "missing email or label id")
		return
	}

	if err := h.Labels.AddToEmail(r.Context(), userID, emailID, labelID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to add label")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) RemoveFromEmail(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	emailID := r.PathValue("id")
	labelID := r.PathValue("labelId")
	if emailID == "" || labelID == "" {
		WriteError(w, http.StatusBadRequest, "missing email or label id")
		return
	}

	if err := h.Labels.RemoveFromEmail(r.Context(), userID, emailID, labelID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to remove label")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) ListByEmail(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	emailID := r.PathValue("id")
	if emailID == "" {
		WriteError(w, http.StatusBadRequest, "missing email id")
		return
	}

	labels, err := h.Labels.ListByEmail(r.Context(), userID, emailID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list labels")
		return
	}
	if labels == nil {
		labels = []domain.Label{}
	}
	WriteJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) BatchAdd(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		EmailIDs []string `json:"EmailIDs"`
		LabelID  string   `json:"LabelID"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.EmailIDs) == 0 || req.LabelID == "" {
		WriteError(w, http.StatusBadRequest, "email_ids and label_id are required")
		return
	}

	if err := h.Labels.BatchAddToEmails(r.Context(), userID, req.EmailIDs, req.LabelID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to add labels")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LabelHandler) BatchRemove(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		EmailIDs []string `json:"EmailIDs"`
		LabelID  string   `json:"LabelID"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.EmailIDs) == 0 || req.LabelID == "" {
		WriteError(w, http.StatusBadRequest, "email_ids and label_id are required")
		return
	}

	if err := h.Labels.BatchRemoveFromEmails(r.Context(), userID, req.EmailIDs, req.LabelID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to remove labels")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
