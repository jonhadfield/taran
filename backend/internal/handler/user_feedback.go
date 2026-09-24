package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/mailer"
)

const (
	maxFeedbackLength = 2000
	// feedbackSendTimeout bounds the mail provider call, which happens before
	// the response: Cloud Run only allocates CPU while a request is in flight.
	feedbackSendTimeout = 10 * time.Second
)

// UserFeedbackHandler passes a message from a user to the site's admins.
type UserFeedbackHandler struct {
	Mailer      mailer.Mailer // nil when no mail provider is configured
	AdminEmails []string
}

func (h *UserFeedbackHandler) Send(w http.ResponseWriter, r *http.Request) {
	if h.Mailer == nil || len(h.AdminEmails) == 0 {
		WriteError(w, http.StatusServiceUnavailable, "feedback is not available right now")
		return
	}

	var req struct {
		Message string `json:"Message"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		WriteError(w, http.StatusBadRequest, "message is required")
		return
	}
	if utf8.RuneCountInString(message) > maxFeedbackLength {
		WriteError(w, http.StatusBadRequest, "message must be 2000 characters or fewer")
		return
	}

	from := auth.UserEmailFromContext(r.Context())
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), feedbackSendTimeout)
	defer cancel()

	// One admin's mail failing shouldn't lose the message for the others.
	delivered := 0
	for _, admin := range h.AdminEmails {
		if err := h.Mailer.SendUserFeedback(ctx, admin, from, message); err != nil {
			slog.Error("failed to send user feedback", "error", err)
			continue
		}
		delivered++
	}

	if delivered == 0 {
		WriteError(w, http.StatusInternalServerError, "could not send your feedback, please try again")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}
