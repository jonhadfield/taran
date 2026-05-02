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

type WaitlistHandler struct {
	Waitlist     database.WaitlistRepository
	Invites      database.InviteRepository
	Mailer       mailer.Mailer // may be nil
	AdminEmails  []string
	AppSettings  *database.AppSettingRepo
}

// Status returns whether the waitlist is open for requests.
// This endpoint is public (no auth required).
func (h *WaitlistHandler) Status(w http.ResponseWriter, r *http.Request) {
	enabled := false
	if h.AppSettings != nil {
		enabled, _ = h.AppSettings.GetBool(r.Context(), "waitlist_enabled", false)
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"waitlistEnabled": enabled})
}

func (h *WaitlistHandler) Submit(w http.ResponseWriter, r *http.Request) {
	// Check if waitlist is enabled
	if h.AppSettings != nil {
		enabled, _ := h.AppSettings.GetBool(r.Context(), "waitlist_enabled", false)
		if !enabled {
			WriteError(w, http.StatusForbidden, "waitlist is not currently open")
			return
		}
	}

	email := strings.ToLower(auth.UserEmailFromContext(r.Context()))
	if email == "" {
		WriteError(w, http.StatusBadRequest, "could not determine user email")
		return
	}

	req := &domain.WaitlistRequest{
		ID:        uuid.New().String(),
		Email:     email,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.Waitlist.Create(r.Context(), req); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to submit waitlist request")
		return
	}

	// Notify admins (best-effort)
	if h.Mailer != nil {
		for _, adminEmail := range h.AdminEmails {
			if err := h.Mailer.SendWaitlistNotification(r.Context(), adminEmail, email); err != nil {
				slog.Error("failed to send waitlist notification", "to", adminEmail, "applicant", email, "error", err)
			}
		}
	}

	WriteJSON(w, http.StatusCreated, req)
}

func (h *WaitlistHandler) List(w http.ResponseWriter, r *http.Request) {
	requests, err := h.Waitlist.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list waitlist requests")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.WaitlistRequest]{Data: requests, Total: len(requests)})
}

func (h *WaitlistHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	adminEmail := auth.UserEmailFromContext(r.Context())

	// Look up the waitlist request
	found, err := h.Waitlist.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to look up waitlist request")
		return
	}
	if found == nil {
		WriteError(w, http.StatusNotFound, "waitlist request not found")
		return
	}

	// Check if invite already exists
	existing, err := h.Invites.GetByEmail(r.Context(), found.Email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to check existing invite")
		return
	}
	if existing != nil {
		// Already invited — just delete waitlist entry
		if err := h.Waitlist.Delete(r.Context(), id); err != nil {
			slog.Error("failed to delete waitlist request after existing invite", "id", id, "error", err)
		}
		WriteJSON(w, http.StatusOK, existing)
		return
	}

	// Create invite
	invite := &domain.Invite{
		ID:        uuid.New().String(),
		Email:     found.Email,
		InvitedBy: adminEmail,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.Invites.Create(r.Context(), invite); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	// Send invite email (best-effort)
	if h.Mailer != nil {
		if err := h.Mailer.SendInvite(r.Context(), found.Email, adminEmail); err != nil {
			slog.Error("failed to send invite email from waitlist approval", "to", found.Email, "error", err)
		}
	}

	// Delete waitlist entry
	if err := h.Waitlist.Delete(r.Context(), id); err != nil {
		slog.Error("failed to delete waitlist request after approval", "id", id, "error", err)
	}

	WriteJSON(w, http.StatusCreated, invite)
}
