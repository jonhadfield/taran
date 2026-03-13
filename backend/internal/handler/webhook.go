package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/email"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/worker"
)

const maxEmailSize = 25 * 1024 * 1024 // 25MB

type WebhookHandler struct {
	Accounts        database.AccountRepository
	Emails          database.EmailRepository
	Extractions     database.ExtractionRepository
	Attachments     database.AttachmentRepository
	WebhookPayloads database.WebhookPayloadRepository
	Resolver        *llm.ProviderResolver
	SenderPrefs     database.SenderPreferenceRepository
	TokenUsage      database.TokenUsageRepository
	Preferences     database.PreferenceRepository
}

func (h *WebhookHandler) IngestEmail(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxEmailSize))
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, "email too large")
		return
	}

	parsed, err := email.Parse(body)
	if err != nil {
		slog.Error("failed to parse email", "error", err)
		WriteError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	toAddress := r.Header.Get("X-Original-To")
	if toAddress == "" {
		toAddress = parsed.To
	}
	toAddress = strings.ToLower(strings.TrimSpace(toAddress))

	account, err := h.Accounts.GetByEmailAddress(r.Context(), toAddress)
	if err != nil {
		slog.Error("failed to look up account", "to", toAddress, "error", err)
		WriteError(w, http.StatusInternalServerError, "account lookup failed")
		return
	}
	if account == nil {
		slog.Warn("no account for address", "to", toAddress)
		WriteError(w, http.StatusNotFound, "unknown recipient")
		return
	}

	if parsed.MessageID != "" {
		existing, _ := h.Emails.GetByMessageID(r.Context(), parsed.MessageID)
		if existing != nil {
			WriteJSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
			return
		}
	}

	now := time.Now()
	emailRecord := &domain.Email{
		ID:          uuid.New().String(),
		UserID:      account.UserID,
		AccountID:   account.ID,
		MessageID:   parsed.MessageID,
		FromAddress: parsed.From,
		FromName:    parsed.FromName,
		ToAddress:   toAddress,
		Subject:     parsed.Subject,
		TextBody:    parsed.TextBody,
		HTMLBody:    parsed.HTMLBody,
		ReceivedAt:  now,
		DateHeader:  parsed.DateHeader,
		Status:            domain.EmailStatusPending,
		UnsubscribeURL:    parsed.UnsubscribeURL,
		UnsubscribeMailto: parsed.UnsubscribeMailto,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.Emails.Create(r.Context(), emailRecord); err != nil {
		slog.Error("failed to store email", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to store email")
		return
	}

	// Store raw webhook payload for replay (best-effort)
	if h.WebhookPayloads != nil {
		payload := &domain.WebhookPayload{
			ID:         uuid.New().String(),
			EmailID:    &emailRecord.ID,
			RawBody:    body,
			Headers:    map[string]string{"X-Original-To": r.Header.Get("X-Original-To")},
			ReceivedAt: now,
			SizeBytes:  len(body),
		}
		if err := h.WebhookPayloads.Create(r.Context(), payload); err != nil {
			slog.Error("failed to store webhook payload", "emailID", emailRecord.ID, "error", err)
		}
	}

	// Store attachment metadata
	if h.Attachments != nil && len(parsed.Attachments) > 0 {
		var attachments []domain.EmailAttachment
		for _, a := range parsed.Attachments {
			attachments = append(attachments, domain.EmailAttachment{
				ID:          uuid.New().String(),
				EmailID:     emailRecord.ID,
				Filename:    a.Filename,
				ContentType: a.ContentType,
				SizeBytes:   a.SizeBytes,
				CreatedAt:   now,
			})
		}
		if err := h.Attachments.CreateBatch(r.Context(), attachments); err != nil {
			slog.Error("failed to store attachments", "emailID", emailRecord.ID, "error", err)
			// Non-fatal: email was already stored, continue
		}
	}

	slog.Info("email ingested",
		"id", emailRecord.ID,
		"from", emailRecord.FromAddress,
		"to", toAddress,
		"subject", emailRecord.Subject,
	)

	// Process extraction synchronously so the summary is available immediately
	if h.Resolver != nil {
		worker.ProcessEmail(r.Context(), emailRecord.ID, h.Emails, h.Extractions, h.Resolver, h.SenderPrefs, h.TokenUsage, h.Preferences)
	}

	WriteJSON(w, http.StatusAccepted, map[string]string{
		"status": "accepted",
		"id":     emailRecord.ID,
	})
}
