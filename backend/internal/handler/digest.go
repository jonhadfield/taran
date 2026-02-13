package handler

import (
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type DigestHandler struct {
	Digests   database.DigestRepository
	Generator *digest.Generator
}

func (h *DigestHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	opts := domain.ListOptions{
		Limit:  intParam(r, "limit", 50),
		Offset: intParam(r, "offset", 0),
	}

	digests, total, err := h.Digests.List(r.Context(), userID, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list digests")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.Digest]{Data: digests, Total: total})
}

func (h *DigestHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	digest, err := h.Digests.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	WriteJSON(w, http.StatusOK, digest)
}

func (h *DigestHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if h.Generator == nil {
		WriteError(w, http.StatusInternalServerError, "digest generation not configured")
		return
	}

	userID := auth.UserIDFromContext(r.Context())

	now := time.Now()
	periodEnd := now
	periodStart := now.Add(-24 * time.Hour)

	d, err := h.Generator.GenerateForUser(r.Context(), userID, "daily", periodStart, periodEnd)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate digest")
		return
	}

	if d == nil {
		WriteError(w, http.StatusUnprocessableEntity, "no emails found in the last 24 hours")
		return
	}

	WriteJSON(w, http.StatusOK, d)
}
