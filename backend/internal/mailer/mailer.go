package mailer

import (
	"context"

	"github.com/hadfielj/taran/backend/internal/domain"
)

type Mailer interface {
	SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error
	SendInvite(ctx context.Context, toEmail, fromName string) error
}
