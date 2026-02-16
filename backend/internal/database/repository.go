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
	SetStatus(ctx context.Context, id string, status domain.EmailStatus, reason string) error
	ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error)
	CountByPeriod(ctx context.Context, userID string, from, to time.Time) (int, error)
	TopSenders(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.SenderCount, error)
}

type ExtractionRepository interface {
	Create(ctx context.Context, extraction *domain.Extraction) error
	GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error)
	ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time) ([]domain.Extraction, error)
	ListTopicsByUser(ctx context.Context, userID string) ([]string, error)
}

type FeedbackRepository interface {
	Upsert(ctx context.Context, fb *domain.EmailFeedback) error
	GetByEmailID(ctx context.Context, userID, emailID string) (*domain.EmailFeedback, error)
	GetSenderStats(ctx context.Context, userID string) ([]domain.SenderFeedbackStat, error)
	GetTopicStats(ctx context.Context, userID string) ([]domain.TopicFeedbackStat, error)
}

type DigestRepository interface {
	Create(ctx context.Context, digest *domain.Digest) error
	GetByID(ctx context.Context, userID, id string) (*domain.Digest, error)
	List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error)
	SetSentAt(ctx context.Context, id string, sentAt time.Time) error
	SetShareToken(ctx context.Context, id, userID, token string) error
	ClearShareToken(ctx context.Context, id, userID string) error
	GetByShareToken(ctx context.Context, token string) (*domain.Digest, error)
}

type AccountRepository interface {
	GetByEmailAddress(ctx context.Context, emailAddress string) (*domain.EmailAccount, error)
	GetByID(ctx context.Context, userID, id string) (*domain.EmailAccount, error)
	ListByUser(ctx context.Context, userID string) ([]domain.EmailAccount, error)
	Create(ctx context.Context, account *domain.EmailAccount) error
	Delete(ctx context.Context, userID, id string) error
}

type PreferenceRepository interface {
	Get(ctx context.Context, userID string) (*domain.UserPreference, error)
	Upsert(ctx context.Context, pref *domain.UserPreference) error
}

type SessionRepository interface {
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

type SenderPreferenceRepository interface {
	Upsert(ctx context.Context, pref *domain.SenderPreference) error
	GetByAddress(ctx context.Context, userID, fromAddress string) (*domain.SenderPreference, error)
	ListByUser(ctx context.Context, userID string) ([]domain.SenderPreference, error)
	ListBlockedAddresses(ctx context.Context, userID string) ([]string, error)
}
