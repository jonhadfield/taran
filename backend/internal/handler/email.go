package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type EmailProcessor interface {
	Enqueue(emailID string)
}

type EmailHandler struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Attachments database.AttachmentRepository
	Processor   EmailProcessor
}

type EmailResponse struct {
	domain.Email
	Extraction  *domain.Extraction      `json:"extraction,omitempty"`
	Attachments []domain.EmailAttachment `json:"attachments,omitempty"`
}

func (h *EmailHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	opts := domain.ListOptions{
		Limit:  clampInt(intParam(r, "limit", 50), 1, 200),
		Offset: clampInt(intParam(r, "offset", 0), 0, 100000),
	}

	if v := r.URL.Query().Get("is_read"); v != "" {
		b := v == "true"
		opts.IsRead = &b
	}
	if v := r.URL.Query().Get("is_starred"); v != "" {
		b := v == "true"
		opts.IsStarred = &b
	}
	if v := r.URL.Query().Get("is_archived"); v != "" {
		b := v == "true"
		opts.IsArchived = &b
	}
	if v := r.URL.Query().Get("search"); v != "" {
		opts.Search = &v
	}
	if v := r.URL.Query().Get("topic"); v != "" {
		opts.Topic = &v
	}
	if v := r.URL.Query().Get("category"); v != "" {
		opts.Category = &v
	}

	emails, total, err := h.Emails.List(r.Context(), userID, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list emails")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.Email]{Data: emails, Total: total})
}

func (h *EmailHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	email, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	resp := EmailResponse{Email: *email}
	extraction, err := h.Extractions.GetByEmailIDScoped(r.Context(), userID, id)
	if err == nil {
		resp.Extraction = extraction
	}
	if h.Attachments != nil {
		attachments, err := h.Attachments.ListByEmailID(r.Context(), id)
		if err == nil && len(attachments) > 0 {
			resp.Attachments = attachments
		}
	}

	WriteJSON(w, http.StatusOK, resp)
}

func (h *EmailHandler) UpdateState(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var state domain.EmailState
	if err := LimitedJSONDecoder(r).Decode(&state); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.Emails.UpdateState(r.Context(), userID, id, state); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update email")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EmailHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	_, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	if err := h.Emails.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete email")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

const maxBatchSize = 100

func (h *EmailHandler) BatchUpdateState(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		IDs   []string         `json:"ids"`
		State domain.EmailState `json:"state"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		WriteError(w, http.StatusBadRequest, "ids required")
		return
	}
	if len(req.IDs) > maxBatchSize {
		WriteError(w, http.StatusBadRequest, "too many ids (max 100)")
		return
	}

	if err := h.Emails.BatchUpdateState(r.Context(), userID, req.IDs, req.State); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update emails")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EmailHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		WriteError(w, http.StatusBadRequest, "ids required")
		return
	}
	if len(req.IDs) > maxBatchSize {
		WriteError(w, http.StatusBadRequest, "too many ids (max 100)")
		return
	}

	if err := h.Emails.BatchDelete(r.Context(), userID, req.IDs); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete emails")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *EmailHandler) Reprocess(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	email, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	if email.Status != domain.EmailStatusFailed && email.Status != domain.EmailStatusSkipped {
		WriteError(w, http.StatusBadRequest, "only failed or skipped emails can be reprocessed")
		return
	}

	if err := h.Extractions.DeleteByEmailIDScoped(r.Context(), userID, id); err != nil {
		slog.Error("failed to delete extraction for reprocess", "emailID", id, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to reprocess email")
		return
	}

	if err := h.Emails.SetStatusScoped(r.Context(), userID, id, domain.EmailStatusPending, ""); err != nil {
		slog.Error("failed to reset status for reprocess", "emailID", id, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to reprocess email")
		return
	}

	h.Processor.Enqueue(id)

	WriteJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}

func clampInt(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func intParam(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

var unsubscribeClient = &http.Client{Timeout: 10 * time.Second}

func (h *EmailHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	email, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	if email.UnsubscribeURL == "" && email.UnsubscribeMailto == "" {
		WriteError(w, http.StatusNotFound, "no unsubscribe option available for this email")
		return
	}

	// Prefer HTTP one-click unsubscribe (RFC 8058)
	if email.UnsubscribeURL != "" {
		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, email.UnsubscribeURL,
			strings.NewReader("List-Unsubscribe=One-Click"))
		if err != nil {
			slog.Error("failed to create unsubscribe request", "url", email.UnsubscribeURL, "error", err)
			// Fall back to redirect
			WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := unsubscribeClient.Do(req)
		if err != nil {
			slog.Warn("one-click unsubscribe POST failed, falling back to redirect", "url", email.UnsubscribeURL, "error", err)
			WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
			return
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			slog.Info("one-click unsubscribe succeeded", "emailID", id, "url", email.UnsubscribeURL)
			WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed", "method": "one-click"})
			return
		}

		// Server rejected the POST — fall back to redirect
		slog.Warn("one-click unsubscribe returned non-success", "status", resp.StatusCode, "url", email.UnsubscribeURL)
		WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
		return
	}

	// Mailto fallback
	WriteJSON(w, http.StatusOK, map[string]string{"status": "mailto", "address": email.UnsubscribeMailto})
}
