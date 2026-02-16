package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type FeedbackHandler struct {
	Feedback database.FeedbackRepository
}

func (h *FeedbackHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	emailID := r.PathValue("id")

	var body struct {
		Rating string `json:"Rating"`
	}
	if err := LimitedJSONDecoder(r).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Rating != "useful" && body.Rating != "not_useful" {
		WriteError(w, http.StatusBadRequest, "rating must be 'useful' or 'not_useful'")
		return
	}

	fb := &domain.EmailFeedback{
		ID:        uuid.New().String(),
		UserID:    userID,
		EmailID:   emailID,
		Rating:    body.Rating,
		CreatedAt: time.Now(),
	}

	if err := h.Feedback.Upsert(r.Context(), fb); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to save feedback")
		return
	}

	WriteJSON(w, http.StatusOK, fb)
}

func (h *FeedbackHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	emailID := r.PathValue("id")

	fb, err := h.Feedback.GetByEmailID(r.Context(), userID, emailID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get feedback")
		return
	}
	if fb == nil {
		WriteJSON(w, http.StatusOK, map[string]any{"Rating": nil})
		return
	}

	WriteJSON(w, http.StatusOK, fb)
}
