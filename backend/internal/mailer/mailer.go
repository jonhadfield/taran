package mailer

import (
	"context"

	"github.com/hadfielj/taran/backend/internal/domain"
)

type Mailer interface {
	SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error
	SendInvite(ctx context.Context, toEmail string) error
	SendInviteApproved(ctx context.Context, toEmail string) error
	SendTokenWarning(ctx context.Context, toEmail string, usagePercent int, tokensUsed, tokenLimit int) error
	SendWaitlistNotification(ctx context.Context, toEmail, applicantEmail string) error
	// SendSignupNotification tells an admin that a new person has got into the
	// app. via describes how, e.g. "open registration" or "an invite".
	SendSignupNotification(ctx context.Context, toEmail, newUserEmail, via string) error
	SendWeeklySummary(ctx context.Context, toEmail string, summary *domain.WeeklySummary, unsubscribeURL string) error
	// SendUserFeedback passes a message from a user to an admin, with the
	// sender's address as reply-to so the admin can simply reply.
	SendUserFeedback(ctx context.Context, toEmail, fromUserEmail, message string) error
}
