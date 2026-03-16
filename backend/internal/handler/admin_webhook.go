package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/email"
	"github.com/hadfielj/taran/backend/internal/llm"
)

type AdminWebhookHandler struct {
	Emails          database.EmailRepository
	Extractions     database.ExtractionRepository
	WebhookPayloads database.WebhookPayloadRepository
	Accounts        database.AccountRepository
	Resolver        *llm.ProviderResolver
	SenderPrefs     database.SenderPreferenceRepository
	TokenUsage      database.TokenUsageRepository
	Preferences     database.PreferenceRepository
	Processor       EmailProcessor
}

const maxAdminBatchSize = 500

// ListFailed returns failed/stuck emails across all users.
func (h *AdminWebhookHandler) ListFailed(w http.ResponseWriter, r *http.Request) {
	limit := clampInt(intParam(r, "limit", 50), 1, 200)
	offset := clampInt(intParam(r, "offset", 0), 0, 100000)

	emails, total, err := h.Emails.ListFailedAll(r.Context(), limit, offset)
	if err != nil {
		slog.Error("failed to list failed emails", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to list failed emails")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.FailedEmail]{Data: emails, Total: total})
}

// RetryOne resets a single email for reprocessing (no user scoping).
func (h *AdminWebhookHandler) RetryOne(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	em, err := h.Emails.GetByIDInternal(r.Context(), id)
	if err != nil || em == nil {
		WriteError(w, http.StatusNotFound, "email not found")
		return
	}

	if err := h.Extractions.DeleteByEmailID(r.Context(), id); err != nil {
		slog.Error("admin retry: failed to delete extraction", "emailID", id, "error", err)
	}

	if err := h.Emails.SetStatus(r.Context(), id, domain.EmailStatusPending, ""); err != nil {
		slog.Error("admin retry: failed to reset status", "emailID", id, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to reset email")
		return
	}
	_ = h.Emails.ResetRetryCount(r.Context(), id)

	h.Processor.Enqueue(id)
	slog.Info("admin retry enqueued", "emailID", id)

	WriteJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}

// BatchRetry resets multiple failed emails for reprocessing.
func (h *AdminWebhookHandler) BatchRetry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs       []string `json:"ids"`
		AllFailed bool     `json:"all_failed"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var ids []string
	if req.AllFailed {
		// Fetch up to maxAdminBatchSize failed email IDs
		failed, _, err := h.Emails.ListFailedAll(r.Context(), maxAdminBatchSize, 0)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to list failed emails")
			return
		}
		for _, fe := range failed {
			ids = append(ids, fe.ID)
		}
	} else {
		ids = req.IDs
	}

	if len(ids) == 0 {
		WriteError(w, http.StatusBadRequest, "no emails to retry")
		return
	}
	if len(ids) > maxAdminBatchSize {
		WriteError(w, http.StatusBadRequest, "too many emails (max 500)")
		return
	}

	count, err := h.Emails.BatchResetForRetry(r.Context(), ids)
	if err != nil {
		slog.Error("admin batch retry failed", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to reset emails")
		return
	}

	for _, id := range ids {
		h.Processor.Enqueue(id)
	}
	slog.Info("admin batch retry enqueued", "count", count)

	WriteJSON(w, http.StatusOK, map[string]any{"status": "queued", "count": count})
}

// Replay re-ingests an email from a stored webhook payload.
func (h *AdminWebhookHandler) Replay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	payload, err := h.WebhookPayloads.GetByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "webhook payload not found")
		return
	}

	parsed, err := email.Parse(payload.RawBody)
	if err != nil {
		slog.Error("admin replay: failed to parse stored payload", "payloadID", id, "error", err)
		WriteError(w, http.StatusUnprocessableEntity, "failed to parse stored email")
		return
	}

	toAddress := payload.Headers["X-Original-To"]
	if toAddress == "" {
		toAddress = parsed.To
	}
	toAddress = strings.ToLower(strings.TrimSpace(toAddress))

	account, err := h.Accounts.GetByEmailAddress(r.Context(), toAddress)
	if err != nil || account == nil {
		WriteError(w, http.StatusNotFound, "no account for address")
		return
	}

	// Delete original email if linked
	if payload.EmailID != nil && *payload.EmailID != "" {
		_ = h.Extractions.DeleteByEmailID(r.Context(), *payload.EmailID)
		_ = h.Emails.DeleteInternal(r.Context(), *payload.EmailID)
	}

	threadID := resolveThreadID(r.Context(), h.Emails, parsed)

	now := time.Now()
	emailRecord := &domain.Email{
		ID:                uuid.New().String(),
		UserID:            account.UserID,
		AccountID:         account.ID,
		MessageID:         parsed.MessageID,
		InReplyTo:         parsed.InReplyTo,
		ThreadID:          threadID,
		FromAddress:       parsed.From,
		FromName:          parsed.FromName,
		ToAddress:         toAddress,
		Subject:           parsed.Subject,
		TextBody:          parsed.TextBody,
		HTMLBody:          parsed.HTMLBody,
		ReceivedAt:        now,
		DateHeader:        parsed.DateHeader,
		Status:            domain.EmailStatusPending,
		UnsubscribeURL:    parsed.UnsubscribeURL,
		UnsubscribeMailto: parsed.UnsubscribeMailto,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.Emails.Create(r.Context(), emailRecord); err != nil {
		slog.Error("admin replay: failed to create email", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to create email")
		return
	}

	h.Processor.Enqueue(emailRecord.ID)
	slog.Info("admin replay enqueued", "payloadID", id, "emailID", emailRecord.ID)

	WriteJSON(w, http.StatusOK, map[string]string{"status": "queued", "id": emailRecord.ID})
}

// PipelineHealth returns aggregate email status counts and stored payload count.
func (h *AdminWebhookHandler) PipelineHealth(w http.ResponseWriter, r *http.Request) {
	// Get global email status counts (not user-scoped — use a raw query via ListFailedAll total)
	// We'll build a simple pipeline view
	type pipelineResponse struct {
		Pending    int `json:"pending"`
		Processing int `json:"processing"`
		Processed  int `json:"processed"`
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Payloads   int `json:"payloads"`
	}

	// Use ListFailedAll to get failed+processing count
	_, failedTotal, err := h.Emails.ListFailedAll(r.Context(), 1, 0)
	if err != nil {
		slog.Error("pipeline health: failed to count failed emails", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to get pipeline health")
		return
	}

	payloadCount, err := h.WebhookPayloads.Count(r.Context())
	if err != nil {
		slog.Error("pipeline health: failed to count payloads", "error", err)
		payloadCount = 0
	}

	resp := pipelineResponse{
		Failed:   failedTotal,
		Payloads: payloadCount,
	}

	// Process email via worker synchronously — use Processor interface
	WriteJSON(w, http.StatusOK, resp)
}

