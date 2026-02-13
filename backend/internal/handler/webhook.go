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
	Accounts    database.AccountRepository
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Provider    llm.Provider
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
		Status:      domain.EmailStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.Emails.Create(r.Context(), emailRecord); err != nil {
		slog.Error("failed to store email", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to store email")
		return
	}

	slog.Info("email ingested",
		"id", emailRecord.ID,
		"from", emailRecord.FromAddress,
		"to", toAddress,
		"subject", emailRecord.Subject,
	)

	// Process extraction synchronously so the summary is available immediately
	if h.Provider != nil {
		worker.ProcessEmail(r.Context(), emailRecord.ID, h.Emails, h.Extractions, h.Provider)
	}

	WriteJSON(w, http.StatusAccepted, map[string]string{
		"status": "accepted",
		"id":     emailRecord.ID,
	})
}
