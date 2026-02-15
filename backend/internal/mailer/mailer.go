package mailer

import (
	"context"

	"github.com/hadfielj/taran/backend/internal/domain"
)

type Mailer interface {
	SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest) error
}
