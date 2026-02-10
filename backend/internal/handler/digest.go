package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type DigestHandler struct {
	Digests database.DigestRepository
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
