package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
)

type TopicHandler struct {
	Extractions database.ExtractionRepository
}

func (h *TopicHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	limit := clampInt(intParam(r, "limit", 15), 5, 50)

	topics, err := h.Extractions.ListTopicsByUser(r.Context(), userID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list topics")
		return
	}
	if topics == nil {
		topics = []string{}
	}

	WriteJSON(w, http.StatusOK, topics)
}
