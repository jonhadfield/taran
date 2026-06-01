package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/mailer"
)

type InviteHandler struct {
	Invites     database.InviteRepository
	AdminEmails []string
	Mailer      mailer.Mailer // may be nil
}

type createInviteRequest struct {
	Email string `json:"email"`
}

func (h *InviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	adminEmail := auth.UserEmailFromContext(r.Context())

	var req createInviteRequest
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		WriteError(w, http.StatusBadRequest, "valid email is required")
		return
	}

	existing, err := h.Invites.GetByEmail(r.Context(), req.Email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check existing invite")
		return
	}
	if existing != nil {
		WriteError(w, http.StatusConflict, "invite already exists for this email")
		return
	}

	now := time.Now().UTC()
	invite := &domain.Invite{
		ID:        uuid.New().String(),
		Email:     req.Email,
		InvitedBy: adminEmail,
		CreatedAt: now,
	}

	if err := h.Invites.Create(r.Context(), invite); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	if h.Mailer != nil {
		if err := h.Mailer.SendInvite(r.Context(), req.Email); err != nil {
			slog.Error("failed to send invite email", "to", req.Email, "error", err)
			// Don't fail the request — invite is created, email delivery is best-effort
		}
	}

	WriteJSON(w, http.StatusCreated, invite)
}

func (h *InviteHandler) List(w http.ResponseWriter, r *http.Request) {
	invites, err := h.Invites.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list invites")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.Invite]{Data: invites, Total: len(invites)})
}

func (h *InviteHandler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(auth.UserEmailFromContext(r.Context()))

	// Admins always have access
	for _, admin := range h.AdminEmails {
		if email == admin {
			WriteJSON(w, http.StatusOK, map[string]any{"hasAccess": true, "reason": "admin"})
			return
		}
	}

	invite, err := h.Invites.GetByEmail(r.Context(), email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check access")
		return
	}

	if invite == nil {
		WriteJSON(w, http.StatusOK, map[string]any{"hasAccess": false, "reason": "not_invited"})
		return
	}

	// Mark accepted on first successful access check
	if invite.AcceptedAt == nil {
		if err := h.Invites.MarkAccepted(r.Context(), email); err != nil {
			slog.Error("failed to mark invite accepted", "email", email, "error", err)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{"hasAccess": true, "reason": "invited"})
}
