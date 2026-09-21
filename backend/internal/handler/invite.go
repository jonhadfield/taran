package handler

import (
	"context"
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
	// Settings enables open registration; nil keeps the app invite-only.
	Settings    auth.BoolSettings
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

// signupNotifyTimeout bounds the sign-up notification so a slow mail
// provider can't hold up a new user's first page load for long.
const signupNotifyTimeout = 5 * time.Second

// notifySignup emails each admin about a new sign-up. It sends before the
// response rather than in the background because Cloud Run only allocates CPU
// while a request is in flight; it happens once per user, and failures are
// only logged.
func (h *InviteHandler) notifySignup(ctx context.Context, email, invitedBy string) {
	if h.Mailer == nil || len(h.AdminEmails) == 0 {
		return
	}
	via := "an invite"
	if invitedBy == auth.OpenRegistrationInviter {
		via = "open registration"
	}
	// Don't let the user navigating away cancel the notification.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), signupNotifyTimeout)
	defer cancel()
	for _, admin := range h.AdminEmails {
		if err := h.Mailer.SendSignupNotification(ctx, admin, email, via); err != nil {
			slog.Error("failed to send sign-up notification", "error", err)
		}
	}
}

func (h *InviteHandler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(auth.UserEmailFromContext(r.Context()))

	checker := auth.AccessChecker{Invites: h.Invites, Settings: h.Settings, AdminEmails: h.AdminEmails}
	access, err := checker.Check(r.Context(), email)
	if err != nil {
		slog.Error("access check failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to check access")
		return
	}

	// Mark accepted on first successful access check. That first acceptance is
	// the moment a new person gets in, so it's when admins are told.
	if access.Invite != nil && access.Invite.AcceptedAt == nil {
		first, err := h.Invites.MarkAccepted(r.Context(), email)
		if err != nil {
			slog.Error("failed to mark invite accepted", "email", email, "error", err)
		} else if first {
			h.notifySignup(r.Context(), email, access.Invite.InvitedBy)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]any{"hasAccess": access.Allowed, "reason": access.Reason})
}
