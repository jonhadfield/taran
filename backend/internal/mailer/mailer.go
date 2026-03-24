package mailer

import (
	"context"

	"github.com/hadfielj/taran/backend/internal/domain"
)

type Mailer interface {
	SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error
	SendInvite(ctx context.Context, toEmail, fromName string) error
	SendTokenWarning(ctx context.Context, toEmail string, usagePercent int, tokensUsed, tokenLimit int) error
	SendWeeklySummary(ctx context.Context, toEmail string, summary *domain.WeeklySummary, unsubscribeURL string) error
}
