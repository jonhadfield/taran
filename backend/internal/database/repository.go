package database

import (
	"context"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
)

type EmailRepository interface {
	Create(ctx context.Context, email *domain.Email) error
	GetByID(ctx context.Context, userID, id string) (*domain.Email, error)
	GetByIDInternal(ctx context.Context, id string) (*domain.Email, error)
	List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error)
	UpdateState(ctx context.Context, userID, id string, state domain.EmailState) error
	GetByMessageID(ctx context.Context, messageID string) (*domain.Email, error)
	ListPending(ctx context.Context, limit int) ([]domain.Email, error)
	SetStatus(ctx context.Context, id string, status domain.EmailStatus) error
	ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error)
}

type ExtractionRepository interface {
	Create(ctx context.Context, extraction *domain.Extraction) error
	GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error)
	ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time) ([]domain.Extraction, error)
}

type DigestRepository interface {
	Create(ctx context.Context, digest *domain.Digest) error
	GetByID(ctx context.Context, userID, id string) (*domain.Digest, error)
	List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error)
}

type AccountRepository interface {
	GetByEmailAddress(ctx context.Context, emailAddress string) (*domain.EmailAccount, error)
	GetByID(ctx context.Context, userID, id string) (*domain.EmailAccount, error)
	ListByUser(ctx context.Context, userID string) ([]domain.EmailAccount, error)
	Create(ctx context.Context, account *domain.EmailAccount) error
	Delete(ctx context.Context, userID, id string) error
}

type SessionRepository interface {
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
}
