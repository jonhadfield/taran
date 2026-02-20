package handler

import (
	"net/http"
	"strconv"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
)

type TopicHandler struct {
	Extractions database.ExtractionRepository
}

func (h *TopicHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	limit := 15
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 5 && n <= 50 {
			limit = n
		}
	}

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
