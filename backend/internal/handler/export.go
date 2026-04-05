package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type ExportHandler struct {
	Emails  database.EmailRepository
	Digests database.DigestRepository
}

type ExportData struct {
	Emails  []domain.Email  `json:"emails"`
	Digests []domain.Digest `json:"digests"`
}

func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	// Fetch emails (up to 1k per export)
	emails, _, err := h.Emails.List(r.Context(), userID, domain.ListOptions{Limit: 1000})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to export emails")
		return
	}

	// Fetch digests (up to 1k per export)
	digests, _, err := h.Digests.List(r.Context(), userID, domain.ListOptions{Limit: 1000})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to export digests")
		return
	}

	if emails == nil {
		emails = []domain.Email{}
	}
	if digests == nil {
		digests = []domain.Digest{}
	}

	w.Header().Set("Content-Disposition", "attachment; filename=mailbrief-export.json")
	WriteJSON(w, http.StatusOK, ExportData{
		Emails:  emails,
		Digests: digests,
	})
}
