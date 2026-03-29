package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

type AccountHandler struct {
	Accounts    database.AccountRepository
	EmailDomain string
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

type createAccountRequest struct {
	Username string `json:"Username"`
}

func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 30 {
		return fmt.Errorf("username must be at most 30 characters")
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	// Enforce single-inbox limit
	existing, err := h.Accounts.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check existing accounts")
		return
	}
	if len(existing) > 0 {
		WriteError(w, http.StatusConflict, "you already have an inbox")
		return
	}

	var req createAccountRequest
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))

	if req.Username == "" {
		WriteError(w, http.StatusBadRequest, "username is required")
		return
	}

	if err := validateUsername(req.Username); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	emailAddress := req.Username + "@" + h.EmailDomain

	taken, err := h.Accounts.GetByEmailAddress(r.Context(), emailAddress)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check username availability")
		return
	}
	if taken != nil {
		WriteError(w, http.StatusConflict, "username is already taken")
		return
	}

	now := time.Now().UTC()
	account := &domain.EmailAccount{
		ID:           uuid.New().String(),
		UserID:       userID,
		EmailAddress: emailAddress,
		DisplayName:  req.Username,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.Accounts.Create(r.Context(), account); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	WriteJSON(w, http.StatusCreated, account)
}

func (h *AccountHandler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	username := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("username")))

	if username == "" {
		WriteError(w, http.StatusBadRequest, "username is required")
		return
	}

	if err := validateUsername(username); err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"available": false, "reason": err.Error()})
		return
	}

	emailAddress := username + "@" + h.EmailDomain

	existing, err := h.Accounts.GetByEmailAddress(r.Context(), emailAddress)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check username availability")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"available": existing == nil})
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.Accounts.Delete(r.Context(), userID, id); err != nil {
		slog.Error("failed to delete account", "accountID", id, "userID", userID, "error", err)
		if err.Error() == "account not found" {
			WriteError(w, http.StatusNotFound, "account not found")
		} else {
			WriteError(w, http.StatusInternalServerError, "failed to delete account")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
