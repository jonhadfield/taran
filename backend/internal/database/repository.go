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
	GetByIDsInternal(ctx context.Context, ids []string) ([]domain.Email, error)
	List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error)
	UpdateState(ctx context.Context, userID, id string, state domain.EmailState) error
	Delete(ctx context.Context, userID, id string) error
	BatchUpdateState(ctx context.Context, userID string, ids []string, state domain.EmailState) error
	BatchDelete(ctx context.Context, userID string, ids []string) error
	GetByMessageID(ctx context.Context, messageID string) (*domain.Email, error)
	ListPending(ctx context.Context, limit int) ([]domain.Email, error)
	SetStatus(ctx context.Context, id string, status domain.EmailStatus, reason string) error
	SetStatusScoped(ctx context.Context, userID, id string, status domain.EmailStatus, reason string) error
	ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error)
	CountByPeriod(ctx context.Context, userID string, from, to time.Time) (int, error)
	CountByWeek(ctx context.Context, userID string, weeks int) ([]domain.WeekCount, error)
	TopSenders(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.SenderCount, error)
	ListSenders(ctx context.Context, userID string) ([]domain.SenderInfo, error)
	CountByStatus(ctx context.Context, userID string) (map[domain.EmailStatus]int, error)
	CountBySenderWeek(ctx context.Context, userID, fromAddress string, weeks int) ([]domain.WeekCount, error)
	GetSenderDetail(ctx context.Context, userID, fromAddress string) (*domain.SenderDetail, error)
	ListRetryable(ctx context.Context, maxRetries, limit int) ([]domain.Email, error)
	IncrementRetryCount(ctx context.Context, id string) error
	ResetRetryCount(ctx context.Context, id string) error
	ListFailedAll(ctx context.Context, limit, offset int) ([]domain.FailedEmail, int, error)
	BatchResetForRetry(ctx context.Context, ids []string) (int, error)
	DeleteInternal(ctx context.Context, id string) error
	CountByHourAndDay(ctx context.Context, userID string) ([]domain.HeatmapCell, error)
	CountByFilter(ctx context.Context, userID string, status *domain.EmailStatus, isRead *bool) (int, error)
	ListSubscriptions(ctx context.Context, userID string) ([]domain.SubscriptionInfo, error)
	GetThreadEmails(ctx context.Context, userID, threadID string) ([]domain.Email, error)
	UpdateThreadID(ctx context.Context, id, threadID string) error
}

type ExtractionRepository interface {
	Create(ctx context.Context, extraction *domain.Extraction) error
	GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error)
	GetByEmailIDScoped(ctx context.Context, userID, emailID string) (*domain.Extraction, error)
	DeleteByEmailID(ctx context.Context, emailID string) error
	DeleteByEmailIDScoped(ctx context.Context, userID, emailID string) error
	ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time, excludedCategories ...string) ([]domain.Extraction, error)
	ListTopicsByUser(ctx context.Context, userID string, limit int) ([]string, error)
	ListTopicsWithCount(ctx context.Context, userID string, limit int) ([]domain.TopicCount, error)
	CountByCategory(ctx context.Context, userID string) ([]domain.CategoryCount, error)
	CountActionItems(ctx context.Context, userID string, from, to time.Time) (int, error)
	GetSummariesByEmailIDs(ctx context.Context, emailIDs []string) (map[string]string, error)
}

type FeedbackRepository interface {
	Upsert(ctx context.Context, fb *domain.EmailFeedback) error
	Delete(ctx context.Context, userID, emailID string) error
	GetByEmailID(ctx context.Context, userID, emailID string) (*domain.EmailFeedback, error)
	GetSenderStats(ctx context.Context, userID string) ([]domain.SenderFeedbackStat, error)
	GetTopicStats(ctx context.Context, userID string) ([]domain.TopicFeedbackStat, error)
}

type DigestRepository interface {
	Create(ctx context.Context, digest *domain.Digest) error
	GetByID(ctx context.Context, userID, id string) (*domain.Digest, error)
	GetByIDInternal(ctx context.Context, id string) (*domain.Digest, error)
	List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error)
	Delete(ctx context.Context, userID, id string) error
	SetSentAt(ctx context.Context, id string, sentAt time.Time) error
	SetShareToken(ctx context.Context, id, userID, token string) error
	ClearShareToken(ctx context.Context, id, userID string) error
	GetByShareToken(ctx context.Context, token string) (*domain.Digest, error)
	ExistsForPeriod(ctx context.Context, userID string, periodStart, periodEnd time.Time) (bool, error)
	ListUnsent(ctx context.Context, olderThan time.Time, limit int) ([]domain.Digest, error)
	DeleteOrphaned(ctx context.Context) (int, error)
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
	SetTokenWarningSent(ctx context.Context, userID string) error
}

type SessionRepository interface {
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	GetUserEmail(ctx context.Context, userID string) (string, error)
	RotateToken(ctx context.Context, oldToken, newToken string) error
}

type InviteRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.Invite, error)
	Create(ctx context.Context, invite *domain.Invite) error
	List(ctx context.Context) ([]domain.Invite, error)
	MarkAccepted(ctx context.Context, email string) error
}

type WaitlistRepository interface {
	Create(ctx context.Context, req *domain.WaitlistRequest) error
	GetByEmail(ctx context.Context, email string) (*domain.WaitlistRequest, error)
	List(ctx context.Context) ([]domain.WaitlistRequest, error)
	Delete(ctx context.Context, id string) error
}

type DigestFeedbackRepository interface {
	Upsert(ctx context.Context, fb *domain.DigestFeedback) error
	Delete(ctx context.Context, userID, digestID string) error
	GetByDigestID(ctx context.Context, userID, digestID string) (*domain.DigestFeedback, error)
}

type AttachmentRepository interface {
	CreateBatch(ctx context.Context, attachments []domain.EmailAttachment) error
	ListByEmailID(ctx context.Context, emailID string) ([]domain.EmailAttachment, error)
}

type TokenUsageRepository interface {
	Create(ctx context.Context, tu *domain.TokenUsage) error
	GetMonthlyTotal(ctx context.Context, userID string) (int, error)
	GetDailyTotal(ctx context.Context, userID string) (int, error)
	GetDailyBreakdown(ctx context.Context, userID string, days int) ([]domain.DailyTokenCount, error)
	GetUsageStats(ctx context.Context, userID string) (*domain.UsageStats, error)
	GetGlobalMonthlyTotal(ctx context.Context) (int, error)
}

type UserLLMKeyRepository interface {
	Upsert(ctx context.Context, key *domain.UserLLMKey) error
	GetByUserAndProvider(ctx context.Context, userID, provider string) (*domain.UserLLMKey, error)
	ListByUser(ctx context.Context, userID string) ([]domain.UserLLMKey, error)
	Update(ctx context.Context, userID, provider string, isActive *bool, model *string) error
	Delete(ctx context.Context, userID, provider string) error
}

type WebhookPayloadRepository interface {
	Create(ctx context.Context, payload *domain.WebhookPayload) error
	GetByID(ctx context.Context, id string) (*domain.WebhookPayload, error)
	GetByEmailID(ctx context.Context, emailID string) (*domain.WebhookPayload, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
	Count(ctx context.Context) (int, error)
}

type LabelRepository interface {
	Create(ctx context.Context, label *domain.Label) error
	GetByID(ctx context.Context, userID, id string) (*domain.Label, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Label, error)
	Update(ctx context.Context, userID string, label *domain.Label) error
	Delete(ctx context.Context, userID, id string) error
	AddToEmail(ctx context.Context, userID, emailID, labelID string) error
	RemoveFromEmail(ctx context.Context, userID, emailID, labelID string) error
	ListByEmail(ctx context.Context, userID, emailID string) ([]domain.Label, error)
	BatchAddToEmails(ctx context.Context, userID string, emailIDs []string, labelID string) error
	BatchRemoveFromEmails(ctx context.Context, userID string, emailIDs []string, labelID string) error
}

type AutoArchiveRuleRepository interface {
	ListByUser(ctx context.Context, userID string) ([]domain.AutoArchiveRule, error)
	Upsert(ctx context.Context, rule *domain.AutoArchiveRule) error
	Delete(ctx context.Context, userID, id string) error
	ListEmailsToArchive(ctx context.Context, limit int) ([]domain.ArchiveCandidate, error)
}

type SavedSearchRepository interface {
	Create(ctx context.Context, s *domain.SavedSearch) error
	ListByUser(ctx context.Context, userID string) ([]domain.SavedSearch, error)
	Delete(ctx context.Context, userID, id string) error
	CountByUser(ctx context.Context, userID string) (int, error)
}

type SenderPreferenceRepository interface {
	Upsert(ctx context.Context, pref *domain.SenderPreference) error
	GetByAddress(ctx context.Context, userID, fromAddress string) (*domain.SenderPreference, error)
	ListByUser(ctx context.Context, userID string) ([]domain.SenderPreference, error)
	ListBlockedAddresses(ctx context.Context, userID string) ([]string, error)
	MarkUnsubscribed(ctx context.Context, userID, fromAddress string) error
}
