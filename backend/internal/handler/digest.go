package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/mailer"
)

type DigestHandler struct {
	Digests           database.DigestRepository
	Generator         *digest.Generator
	Preferences       database.PreferenceRepository
	Mailer            mailer.Mailer
	Sessions          database.SessionRepository
	BaseURL           string
	UnsubscribeSecret string
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

	d, err := h.Digests.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	WriteJSON(w, http.StatusOK, d)
}

func (h *DigestHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if h.Generator == nil {
		WriteError(w, http.StatusInternalServerError, "digest generation not configured")
		return
	}

	userID := auth.UserIDFromContext(r.Context())

	// Determine period based on user's frequency preference
	periodType := "daily"
	if h.Preferences != nil {
		pref, err := h.Preferences.Get(r.Context(), userID)
		if err == nil && pref.DigestFrequency == "weekly" {
			periodType = "weekly"
		}
	}

	now := time.Now()
	periodEnd := now
	periodStart := now.Add(-24 * time.Hour)
	if periodType == "weekly" {
		periodStart = now.Add(-7 * 24 * time.Hour)
	}

	if v := r.URL.Query().Get("period_start"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid period_start format, use RFC3339")
			return
		}
		periodStart = t
	}
	if v := r.URL.Query().Get("period_end"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid period_end format, use RFC3339")
			return
		}
		periodEnd = t
	}

	if !periodEnd.After(periodStart) {
		WriteError(w, http.StatusBadRequest, "period_end must be after period_start")
		return
	}
	if periodEnd.Sub(periodStart) > 30*24*time.Hour {
		WriteError(w, http.StatusBadRequest, "date range cannot exceed 30 days")
		return
	}

	d, err := h.Generator.GenerateForUser(r.Context(), userID, periodType, periodStart, periodEnd)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate digest")
		return
	}

	if d == nil {
		WriteError(w, http.StatusUnprocessableEntity, "no emails found in the selected period")
		return
	}

	WriteJSON(w, http.StatusOK, d)
}

func (h *DigestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	_, err := h.Digests.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	if err := h.Digests.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete digest")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *DigestHandler) Share(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	// Generate a random share token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	if err := h.Digests.SetShareToken(r.Context(), id, userID, token); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to share digest")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ShareToken": token})
}

func (h *DigestHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.Digests.ClearShareToken(r.Context(), id, userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to unshare digest")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DigestHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if h.Mailer == nil {
		WriteError(w, http.StatusServiceUnavailable, "email delivery not configured")
		return
	}

	id := r.PathValue("id")

	d, err := h.Digests.GetByIDInternal(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	email, err := h.Sessions.GetUserEmail(r.Context(), d.UserID)
	if err != nil {
		slog.Error("failed to get user email for digest send", "userID", d.UserID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to look up user email")
		return
	}

	// Auto-generate share token if missing
	if d.ShareToken == nil {
		tokenBytes := make([]byte, 16)
		if _, err := rand.Read(tokenBytes); err != nil {
			slog.Error("failed to generate share token", "digestID", d.ID, "error", err)
		} else {
			token := hex.EncodeToString(tokenBytes)
			if err := h.Digests.SetShareToken(r.Context(), d.ID, d.UserID, token); err != nil {
				slog.Error("failed to set share token", "digestID", d.ID, "error", err)
			} else {
				d.ShareToken = &token
			}
		}
	}

	// Populate transient EmailSummaries from persisted Items for email rendering
	for _, item := range d.Items {
		d.EmailSummaries = append(d.EmailSummaries, domain.DigestEmailSummary{
			EmailID:    item.EmailID,
			Subject:    item.Subject,
			SenderName: item.FromName,
			Summary:    item.Summary,
		})
	}

	var unsubURL string
	if h.BaseURL != "" && h.UnsubscribeSecret != "" {
		unsubURL = mailer.GenerateUnsubscribeURL(h.BaseURL, d.UserID, h.UnsubscribeSecret)
	}

	if err := h.Mailer.SendDigest(r.Context(), email, "", d, unsubURL); err != nil {
		slog.Error("failed to send digest email", "digestID", d.ID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to send digest email")
		return
	}

	now := time.Now()
	if err := h.Digests.SetSentAt(r.Context(), d.ID, now); err != nil {
		slog.Error("failed to set digest sent_at", "digestID", d.ID, "error", err)
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "sent", "to": email})
}

func (h *DigestHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		WriteError(w, http.StatusBadRequest, "missing token")
		return
	}

	d, err := h.Digests.GetByShareToken(r.Context(), token)
	if err != nil || d == nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	// Strip sensitive fields for public view
	d.UserID = ""

	WriteJSON(w, http.StatusOK, d)
}
