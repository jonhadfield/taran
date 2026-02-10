package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type AccountHandler struct {
	Accounts database.AccountRepository
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	accounts, err := h.Accounts.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list accounts")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.EmailAccount]{Data: accounts, Total: len(accounts)})
}
