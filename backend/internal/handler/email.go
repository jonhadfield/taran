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
	"github.com/hadfielj/taran/backend/internal/webhook"
)

type EmailProcessor interface {
	Enqueue(emailID string)
}

type EmailHandler struct {
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Attachments database.AttachmentRepository
	Processor   EmailProcessor
	SenderPrefs database.SenderPreferenceRepository
}

const (
	defaultLimit = 50
	maxLimit     = 200
	maxOffset    = 100000
)

type EmailResponse struct {
	domain.Email
	Extraction  *domain.Extraction      `json:"Extraction,omitempty"`
	Attachments []domain.EmailAttachment `json:"Attachments,omitempty"`
	ThreadCount int                      `json:"ThreadCount,omitempty"`
}

func (h *EmailHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	opts := domain.ListOptions{
		Limit:  clampInt(intParam(r, "limit", defaultLimit), 1, maxLimit),
		Offset: clampInt(intParam(r, "offset", 0), 0, maxOffset),
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
	if v := r.URL.Query().Get("from_address"); v != "" {
		opts.FromAddress = &v
	}
	if v := r.URL.Query().Get("label"); v != "" {
		opts.LabelID = &v
	}
	if v := r.URL.Query().Get("has_attachment"); v == "true" {
		b := true
		opts.HasAttachment = &b
	}
	if v := r.URL.Query().Get("sort"); v == "oldest" || v == "relevance" {
		opts.Sort = v
	}
	if v := r.URL.Query().Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Since = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			opts.Since = &t
		}
	}
	if v := r.URL.Query().Get("before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			opts.Before = &t
		} else if t, err := time.Parse("2006-01-02", v); err == nil {
			opts.Before = &t
		}
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
		attachments, err := h.Attachments.ListByEmailID(r.Context(), id, userID)
		if err == nil && len(attachments) > 0 {
			resp.Attachments = attachments
		}
	}
	// Include thread count if this email is part of a thread
	if email.ThreadID != "" {
		thread, err := h.Emails.GetThreadEmails(r.Context(), userID, email.ThreadID)
		if err == nil {
			resp.ThreadCount = len(thread)
		}
	}

	WriteJSON(w, http.StatusOK, resp)
}

type ThreadEmail struct {
	ID          string `json:"ID"`
	Subject     string `json:"Subject"`
	FromName    string `json:"FromName"`
	FromAddress string `json:"FromAddress"`
	ReceivedAt  string `json:"ReceivedAt"`
	IsRead      bool   `json:"IsRead"`
	Summary     string `json:"Summary,omitempty"`
}

func (h *EmailHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	email, err := h.Emails.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	if email.ThreadID == "" {
		WriteJSON(w, http.StatusOK, []ThreadEmail{})
		return
	}

	thread, err := h.Emails.GetThreadEmails(r.Context(), userID, email.ThreadID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get thread")
		return
	}

	// Batch-fetch extraction summaries for all thread emails (avoids N+1)
	emailIDs := make([]string, len(thread))
	for i, e := range thread {
		emailIDs[i] = e.ID
	}
	summaries, _ := h.Extractions.GetSummariesByEmailIDs(r.Context(), emailIDs)

	result := make([]ThreadEmail, 0, len(thread))
	for _, e := range thread {
		te := ThreadEmail{
			ID:          e.ID,
			Subject:     e.Subject,
			FromName:    e.FromName,
			FromAddress: e.FromAddress,
			ReceivedAt:  e.ReceivedAt.Format("2006-01-02T15:04:05Z07:00"),
			IsRead:      e.IsRead,
			Summary:     summaries[e.ID],
		}
		result = append(result, te)
	}

	WriteJSON(w, http.StatusOK, result)
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

	w.WriteHeader(http.StatusNoContent)
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

	// Reset retry count so manual reprocess starts fresh
	_ = h.Emails.ResetRetryCount(r.Context(), id)

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

// unsubscribeClient uses SafeHTTPClient to prevent SSRF via DNS rebinding.
// Validates resolved IPs at connection time, not just pre-flight.
var unsubscribeClient = webhook.SafeHTTPClient()

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

	// Prefer HTTP one-click unsubscribe, but only where the sender advertised
	// List-Unsubscribe-Post (RFC 8058). Without that header an automated POST
	// is unsolicited, and turns this endpoint into a way to make the server
	// issue attacker-chosen POST requests from its own address.
	if email.UnsubscribeURL != "" && !email.UnsubscribePost {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
		return
	}

	if email.UnsubscribeURL != "" {
		// SSRF protection: validate URL before making outbound request
		if err := webhook.ValidateExternalURL(email.UnsubscribeURL); err != nil {
			slog.Warn("unsubscribe URL blocked by SSRF filter", "url", email.UnsubscribeURL, "error", err)
			WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, email.UnsubscribeURL,
			strings.NewReader("List-Unsubscribe=One-Click"))
		if err != nil {
			slog.Error("failed to create unsubscribe request", "url", email.UnsubscribeURL, "error", err)
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
			if h.SenderPrefs != nil {
				_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, email.FromAddress)
			}
			WriteJSON(w, http.StatusOK, map[string]string{"status": "unsubscribed", "method": "one-click"})
			return
		}

		slog.Warn("one-click unsubscribe returned non-success", "status", resp.StatusCode, "url", email.UnsubscribeURL)
		WriteJSON(w, http.StatusOK, map[string]string{"status": "redirect", "url": email.UnsubscribeURL})
		return
	}

	// Mailto fallback — also track it since user initiated the action
	if h.SenderPrefs != nil {
		_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, email.FromAddress)
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "mailto", "address": email.UnsubscribeMailto})
}

func (h *EmailHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	subs, err := h.Emails.ListSubscriptions(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}
	if subs == nil {
		subs = []domain.SubscriptionInfo{}
	}

	// Merge unsubscribed_at from sender preferences
	if h.SenderPrefs != nil {
		prefs, err := h.SenderPrefs.ListByUser(r.Context(), userID)
		if err == nil {
			prefMap := make(map[string]*time.Time, len(prefs))
			for _, p := range prefs {
				if p.UnsubscribedAt != nil {
					t := *p.UnsubscribedAt
					prefMap[p.FromAddress] = &t
				}
			}
			for i := range subs {
				if t, ok := prefMap[subs[i].FromAddress]; ok {
					subs[i].UnsubscribedAt = t
				}
			}
		}
	}

	WriteJSON(w, http.StatusOK, subs)
}

func (h *EmailHandler) BatchUnsubscribe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req struct {
		FromAddresses []string `json:"FromAddresses"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.FromAddresses) == 0 {
		WriteError(w, http.StatusBadRequest, "no addresses provided")
		return
	}
	if len(req.FromAddresses) > 50 {
		WriteError(w, http.StatusBadRequest, "maximum 50 addresses per batch")
		return
	}

	// Fetch subscriptions once (not per-address) to avoid N+1 queries
	subs, err := h.Emails.ListSubscriptions(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}
	subMap := make(map[string]*domain.SubscriptionInfo, len(subs))
	for i := range subs {
		subMap[subs[i].FromAddress] = &subs[i]
	}

	results := make([]map[string]string, 0, len(req.FromAddresses))
	for _, addr := range req.FromAddresses {
		sub := subMap[addr]
		if sub == nil || (sub.UnsubscribeURL == "" && sub.UnsubscribeMailto == "") {
			results = append(results, map[string]string{"address": addr, "status": "skipped", "reason": "no unsubscribe link"})
			continue
		}

		// Only senders advertising RFC 8058 List-Unsubscribe-Post get an
		// automated POST; everything else is handed back as a redirect.
		if sub.UnsubscribeURL != "" && !sub.UnsubscribePost {
			if h.SenderPrefs != nil {
				_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, addr)
			}
			results = append(results, map[string]string{"address": addr, "status": "redirect", "url": sub.UnsubscribeURL})
			continue
		}

		// Try one-click unsubscribe
		if sub.UnsubscribeURL != "" {
			// SSRF protection
			if err := webhook.ValidateExternalURL(sub.UnsubscribeURL); err != nil {
				slog.Warn("batch unsubscribe URL blocked by SSRF filter", "address", addr, "url", sub.UnsubscribeURL, "error", err)
				results = append(results, map[string]string{"address": addr, "status": "redirect", "url": sub.UnsubscribeURL})
				continue
			}
			httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, sub.UnsubscribeURL,
				strings.NewReader("List-Unsubscribe=One-Click"))
			if err == nil {
				httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := unsubscribeClient.Do(httpReq)
				if err == nil {
					resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						if h.SenderPrefs != nil {
							_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, addr)
						}
						results = append(results, map[string]string{"address": addr, "status": "unsubscribed"})
						continue
					}
				}
			}
			// One-click failed — mark as redirect
			if h.SenderPrefs != nil {
				_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, addr)
			}
			results = append(results, map[string]string{"address": addr, "status": "redirect", "url": sub.UnsubscribeURL})
			continue
		}

		// Mailto only — mark as unsubscribed since user requested it
		if h.SenderPrefs != nil {
			_ = h.SenderPrefs.MarkUnsubscribed(r.Context(), userID, addr)
		}
		results = append(results, map[string]string{"address": addr, "status": "mailto", "address_mailto": sub.UnsubscribeMailto})
	}

	WriteJSON(w, http.StatusOK, results)
}
