package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
)

// SetStatusCall records a call to SetStatus for assertion in tests.
type SetStatusCall struct {
	ID     string
	Status domain.EmailStatus
	Reason string
}

// MockEmailRepo implements database.EmailRepository for testing.
type MockEmailRepo struct {
	CreateFn            func(ctx context.Context, email *domain.Email) error
	GetByIDFn           func(ctx context.Context, userID, id string) (*domain.Email, error)
	GetByIDInternalFn   func(ctx context.Context, id string) (*domain.Email, error)
	GetByIDsInternalFn  func(ctx context.Context, ids []string) ([]domain.Email, error)
	ListFn              func(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error)
	UpdateStateFn       func(ctx context.Context, userID, id string, state domain.EmailState) error
	DeleteFn            func(ctx context.Context, userID, id string) error
	GetByMessageIDFn    func(ctx context.Context, messageID string) (*domain.Email, error)
	ListPendingFn       func(ctx context.Context, limit int) ([]domain.Email, error)
	SetStatusFn         func(ctx context.Context, id string, status domain.EmailStatus, reason string) error
	ListActiveUserIDsFn func(ctx context.Context, from, to time.Time) ([]string, error)
	CountByPeriodFn     func(ctx context.Context, userID string, from, to time.Time) (int, error)
	CountByWeekFn       func(ctx context.Context, userID string, weeks int) ([]domain.WeekCount, error)
	TopSendersFn        func(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.SenderCount, error)
	ListSendersFn       func(ctx context.Context, userID string) ([]domain.SenderInfo, error)
	CountByStatusFn       func(ctx context.Context, userID string) (map[domain.EmailStatus]int, error)
	CountBySenderWeekFn   func(ctx context.Context, userID, fromAddress string, weeks int) ([]domain.WeekCount, error)
	BatchUpdateStateFn    func(ctx context.Context, userID string, ids []string, state domain.EmailState) error
	BatchDeleteFn         func(ctx context.Context, userID string, ids []string) error

	mu             sync.Mutex
	SetStatusCalls []SetStatusCall
	CreateCalls    []*domain.Email
}

func (m *MockEmailRepo) Create(ctx context.Context, email *domain.Email) error {
	m.mu.Lock()
	m.CreateCalls = append(m.CreateCalls, email)
	m.mu.Unlock()
	if m.CreateFn != nil {
		return m.CreateFn(ctx, email)
	}
	return nil
}

func (m *MockEmailRepo) GetByID(ctx context.Context, userID, id string) (*domain.Email, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockEmailRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Email, error) {
	if m.GetByIDInternalFn != nil {
		return m.GetByIDInternalFn(ctx, id)
	}
	return nil, nil
}

func (m *MockEmailRepo) GetByIDsInternal(ctx context.Context, ids []string) ([]domain.Email, error) {
	if m.GetByIDsInternalFn != nil {
		return m.GetByIDsInternalFn(ctx, ids)
	}
	// Fallback: call GetByIDInternal per item for backward compat with existing tests
	var result []domain.Email
	for _, id := range ids {
		e, err := m.GetByIDInternal(ctx, id)
		if err != nil || e == nil {
			continue
		}
		result = append(result, *e)
	}
	return result, nil
}

func (m *MockEmailRepo) List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID, opts)
	}
	return nil, 0, nil
}

func (m *MockEmailRepo) UpdateState(ctx context.Context, userID, id string, state domain.EmailState) error {
	if m.UpdateStateFn != nil {
		return m.UpdateStateFn(ctx, userID, id, state)
	}
	return nil
}

func (m *MockEmailRepo) Delete(ctx context.Context, userID, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, id)
	}
	return nil
}

func (m *MockEmailRepo) CountByWeek(ctx context.Context, userID string, weeks int) ([]domain.WeekCount, error) {
	if m.CountByWeekFn != nil {
		return m.CountByWeekFn(ctx, userID, weeks)
	}
	return nil, nil
}

func (m *MockEmailRepo) GetByMessageID(ctx context.Context, messageID string) (*domain.Email, error) {
	if m.GetByMessageIDFn != nil {
		return m.GetByMessageIDFn(ctx, messageID)
	}
	return nil, nil
}

func (m *MockEmailRepo) ListPending(ctx context.Context, limit int) ([]domain.Email, error) {
	if m.ListPendingFn != nil {
		return m.ListPendingFn(ctx, limit)
	}
	return nil, nil
}

func (m *MockEmailRepo) SetStatus(ctx context.Context, id string, status domain.EmailStatus, reason string) error {
	m.mu.Lock()
	m.SetStatusCalls = append(m.SetStatusCalls, SetStatusCall{ID: id, Status: status, Reason: reason})
	m.mu.Unlock()
	if m.SetStatusFn != nil {
		return m.SetStatusFn(ctx, id, status, reason)
	}
	return nil
}

func (m *MockEmailRepo) SetStatusScoped(ctx context.Context, userID, id string, status domain.EmailStatus, reason string) error {
	m.mu.Lock()
	m.SetStatusCalls = append(m.SetStatusCalls, SetStatusCall{ID: id, Status: status, Reason: reason})
	m.mu.Unlock()
	if m.SetStatusFn != nil {
		return m.SetStatusFn(ctx, id, status, reason)
	}
	return nil
}

func (m *MockEmailRepo) ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error) {
	if m.ListActiveUserIDsFn != nil {
		return m.ListActiveUserIDsFn(ctx, from, to)
	}
	return nil, nil
}

func (m *MockEmailRepo) CountByPeriod(ctx context.Context, userID string, from, to time.Time) (int, error) {
	if m.CountByPeriodFn != nil {
		return m.CountByPeriodFn(ctx, userID, from, to)
	}
	return 0, nil
}

func (m *MockEmailRepo) TopSenders(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.SenderCount, error) {
	if m.TopSendersFn != nil {
		return m.TopSendersFn(ctx, userID, from, to, limit)
	}
	return nil, nil
}

func (m *MockEmailRepo) ListSenders(ctx context.Context, userID string) ([]domain.SenderInfo, error) {
	if m.ListSendersFn != nil {
		return m.ListSendersFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockEmailRepo) CountByStatus(ctx context.Context, userID string) (map[domain.EmailStatus]int, error) {
	if m.CountByStatusFn != nil {
		return m.CountByStatusFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockEmailRepo) BatchUpdateState(ctx context.Context, userID string, ids []string, state domain.EmailState) error {
	if m.BatchUpdateStateFn != nil {
		return m.BatchUpdateStateFn(ctx, userID, ids, state)
	}
	return nil
}

func (m *MockEmailRepo) BatchDelete(ctx context.Context, userID string, ids []string) error {
	if m.BatchDeleteFn != nil {
		return m.BatchDeleteFn(ctx, userID, ids)
	}
	return nil
}

func (m *MockEmailRepo) CountBySenderWeek(ctx context.Context, userID, fromAddress string, weeks int) ([]domain.WeekCount, error) {
	if m.CountBySenderWeekFn != nil {
		return m.CountBySenderWeekFn(ctx, userID, fromAddress, weeks)
	}
	return nil, nil
}

func (m *MockEmailRepo) GetSenderDetail(ctx context.Context, userID, fromAddress string) (*domain.SenderDetail, error) {
	return nil, nil
}

func (m *MockEmailRepo) ListRetryable(ctx context.Context, maxRetries, limit int) ([]domain.Email, error) {
	return nil, nil
}

func (m *MockEmailRepo) IncrementRetryCount(ctx context.Context, id string) error {
	return nil
}

func (m *MockEmailRepo) ResetRetryCount(ctx context.Context, id string) error {
	return nil
}

func (m *MockEmailRepo) ListFailedAll(ctx context.Context, limit, offset int) ([]domain.FailedEmail, int, error) {
	return nil, 0, nil
}

func (m *MockEmailRepo) BatchResetForRetry(ctx context.Context, ids []string) (int, error) {
	return 0, nil
}

func (m *MockEmailRepo) DeleteInternal(ctx context.Context, id string) error {
	return nil
}

func (m *MockEmailRepo) CountByHourAndDay(ctx context.Context, userID string) ([]domain.HeatmapCell, error) {
	return nil, nil
}

func (m *MockEmailRepo) ListSubscriptions(ctx context.Context, userID string) ([]domain.SubscriptionInfo, error) {
	return nil, nil
}

func (m *MockEmailRepo) GetThreadEmails(ctx context.Context, userID, threadID string) ([]domain.Email, error) {
	return nil, nil
}

func (m *MockEmailRepo) UpdateThreadID(ctx context.Context, id, threadID string) error {
	return nil
}

// MockExtractionRepo implements database.ExtractionRepository for testing.
type MockExtractionRepo struct {
	CreateFn                func(ctx context.Context, extraction *domain.Extraction) error
	GetByEmailIDFn          func(ctx context.Context, emailID string) (*domain.Extraction, error)
	GetByEmailIDScopedFn    func(ctx context.Context, userID, emailID string) (*domain.Extraction, error)
	DeleteByEmailIDFn       func(ctx context.Context, emailID string) error
	DeleteByEmailIDScopedFn func(ctx context.Context, userID, emailID string) error
	ListByUserAndPeriodFn   func(ctx context.Context, userID string, from, to time.Time, excludedCategories ...string) ([]domain.Extraction, error)
	ListTopicsByUserFn      func(ctx context.Context, userID string, limit int) ([]string, error)
}

func (m *MockExtractionRepo) Create(ctx context.Context, extraction *domain.Extraction) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, extraction)
	}
	return nil
}

func (m *MockExtractionRepo) GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error) {
	if m.GetByEmailIDFn != nil {
		return m.GetByEmailIDFn(ctx, emailID)
	}
	return nil, nil
}

func (m *MockExtractionRepo) GetByEmailIDScoped(ctx context.Context, userID, emailID string) (*domain.Extraction, error) {
	if m.GetByEmailIDScopedFn != nil {
		return m.GetByEmailIDScopedFn(ctx, userID, emailID)
	}
	if m.GetByEmailIDFn != nil {
		return m.GetByEmailIDFn(ctx, emailID)
	}
	return nil, nil
}

func (m *MockExtractionRepo) DeleteByEmailID(ctx context.Context, emailID string) error {
	if m.DeleteByEmailIDFn != nil {
		return m.DeleteByEmailIDFn(ctx, emailID)
	}
	return nil
}

func (m *MockExtractionRepo) DeleteByEmailIDScoped(ctx context.Context, userID, emailID string) error {
	if m.DeleteByEmailIDScopedFn != nil {
		return m.DeleteByEmailIDScopedFn(ctx, userID, emailID)
	}
	if m.DeleteByEmailIDFn != nil {
		return m.DeleteByEmailIDFn(ctx, emailID)
	}
	return nil
}

func (m *MockExtractionRepo) ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time, excludedCategories ...string) ([]domain.Extraction, error) {
	if m.ListByUserAndPeriodFn != nil {
		return m.ListByUserAndPeriodFn(ctx, userID, from, to, excludedCategories...)
	}
	return nil, nil
}

func (m *MockExtractionRepo) ListTopicsByUser(ctx context.Context, userID string, limit int) ([]string, error) {
	if m.ListTopicsByUserFn != nil {
		return m.ListTopicsByUserFn(ctx, userID, limit)
	}
	return nil, nil
}

func (m *MockExtractionRepo) ListTopicsWithCount(ctx context.Context, userID string, limit int) ([]domain.TopicCount, error) {
	return nil, nil
}

func (m *MockExtractionRepo) CountByCategory(ctx context.Context, userID string) ([]domain.CategoryCount, error) {
	return nil, nil
}

func (m *MockExtractionRepo) CountActionItems(ctx context.Context, userID string, from, to time.Time) (int, error) {
	return 0, nil
}

// MockFeedbackRepo implements database.FeedbackRepository for testing.
type MockFeedbackRepo struct {
	UpsertFn          func(ctx context.Context, fb *domain.EmailFeedback) error
	DeleteFn          func(ctx context.Context, userID, emailID string) error
	GetByEmailIDFn    func(ctx context.Context, userID, emailID string) (*domain.EmailFeedback, error)
	GetSenderStatsFn  func(ctx context.Context, userID string) ([]domain.SenderFeedbackStat, error)
	GetTopicStatsFn   func(ctx context.Context, userID string) ([]domain.TopicFeedbackStat, error)
}

func (m *MockFeedbackRepo) Upsert(ctx context.Context, fb *domain.EmailFeedback) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, fb)
	}
	return nil
}

func (m *MockFeedbackRepo) Delete(ctx context.Context, userID, emailID string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, emailID)
	}
	return nil
}

func (m *MockFeedbackRepo) GetByEmailID(ctx context.Context, userID, emailID string) (*domain.EmailFeedback, error) {
	if m.GetByEmailIDFn != nil {
		return m.GetByEmailIDFn(ctx, userID, emailID)
	}
	return nil, nil
}

func (m *MockFeedbackRepo) GetSenderStats(ctx context.Context, userID string) ([]domain.SenderFeedbackStat, error) {
	if m.GetSenderStatsFn != nil {
		return m.GetSenderStatsFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockFeedbackRepo) GetTopicStats(ctx context.Context, userID string) ([]domain.TopicFeedbackStat, error) {
	if m.GetTopicStatsFn != nil {
		return m.GetTopicStatsFn(ctx, userID)
	}
	return nil, nil
}

// MockDigestRepo implements database.DigestRepository for testing.
type MockDigestRepo struct {
	CreateFn            func(ctx context.Context, digest *domain.Digest) error
	GetByIDFn           func(ctx context.Context, userID, id string) (*domain.Digest, error)
	GetByIDInternalFn   func(ctx context.Context, id string) (*domain.Digest, error)
	ListFn              func(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error)
	DeleteFn            func(ctx context.Context, userID, id string) error
	SetSentAtFn         func(ctx context.Context, id string, sentAt time.Time) error
	SetShareTokenFn     func(ctx context.Context, id, userID, token string) error
	ClearShareTokenFn   func(ctx context.Context, id, userID string) error
	GetByShareTokenFn   func(ctx context.Context, token string) (*domain.Digest, error)
	ExistsForPeriodFn   func(ctx context.Context, userID string, periodStart, periodEnd time.Time) (bool, error)
	ListUnsentFn        func(ctx context.Context, olderThan time.Time, limit int) ([]domain.Digest, error)
}

func (m *MockDigestRepo) Create(ctx context.Context, digest *domain.Digest) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, digest)
	}
	return nil
}

func (m *MockDigestRepo) GetByID(ctx context.Context, userID, id string) (*domain.Digest, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockDigestRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Digest, error) {
	if m.GetByIDInternalFn != nil {
		return m.GetByIDInternalFn(ctx, id)
	}
	return nil, nil
}

func (m *MockDigestRepo) List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID, opts)
	}
	return nil, 0, nil
}

func (m *MockDigestRepo) Delete(ctx context.Context, userID, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, id)
	}
	return nil
}

func (m *MockDigestRepo) SetSentAt(ctx context.Context, id string, sentAt time.Time) error {
	if m.SetSentAtFn != nil {
		return m.SetSentAtFn(ctx, id, sentAt)
	}
	return nil
}

func (m *MockDigestRepo) SetShareToken(ctx context.Context, id, userID, token string) error {
	if m.SetShareTokenFn != nil {
		return m.SetShareTokenFn(ctx, id, userID, token)
	}
	return nil
}

func (m *MockDigestRepo) ClearShareToken(ctx context.Context, id, userID string) error {
	if m.ClearShareTokenFn != nil {
		return m.ClearShareTokenFn(ctx, id, userID)
	}
	return nil
}

func (m *MockDigestRepo) GetByShareToken(ctx context.Context, token string) (*domain.Digest, error) {
	if m.GetByShareTokenFn != nil {
		return m.GetByShareTokenFn(ctx, token)
	}
	return nil, nil
}

func (m *MockDigestRepo) ExistsForPeriod(ctx context.Context, userID string, periodStart, periodEnd time.Time) (bool, error) {
	if m.ExistsForPeriodFn != nil {
		return m.ExistsForPeriodFn(ctx, userID, periodStart, periodEnd)
	}
	return false, nil
}

func (m *MockDigestRepo) ListUnsent(ctx context.Context, olderThan time.Time, limit int) ([]domain.Digest, error) {
	if m.ListUnsentFn != nil {
		return m.ListUnsentFn(ctx, olderThan, limit)
	}
	return nil, nil
}

// MockAccountRepo implements database.AccountRepository for testing.
type MockAccountRepo struct {
	GetByEmailAddressFn func(ctx context.Context, emailAddress string) (*domain.EmailAccount, error)
	GetByIDFn           func(ctx context.Context, userID, id string) (*domain.EmailAccount, error)
	ListByUserFn        func(ctx context.Context, userID string) ([]domain.EmailAccount, error)
	CreateFn            func(ctx context.Context, account *domain.EmailAccount) error
	DeleteFn            func(ctx context.Context, userID, id string) error
}

func (m *MockAccountRepo) GetByEmailAddress(ctx context.Context, emailAddress string) (*domain.EmailAccount, error) {
	if m.GetByEmailAddressFn != nil {
		return m.GetByEmailAddressFn(ctx, emailAddress)
	}
	return nil, nil
}

func (m *MockAccountRepo) GetByID(ctx context.Context, userID, id string) (*domain.EmailAccount, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, userID, id)
	}
	return nil, nil
}

func (m *MockAccountRepo) ListByUser(ctx context.Context, userID string) ([]domain.EmailAccount, error) {
	if m.ListByUserFn != nil {
		return m.ListByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockAccountRepo) Create(ctx context.Context, account *domain.EmailAccount) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, account)
	}
	return nil
}

func (m *MockAccountRepo) Delete(ctx context.Context, userID, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, id)
	}
	return nil
}

// MockSessionRepo implements database.SessionRepository for testing.
type MockSessionRepo struct {
	GetByTokenFn   func(ctx context.Context, token string) (*domain.Session, error)
	GetUserEmailFn func(ctx context.Context, userID string) (string, error)
}

func (m *MockSessionRepo) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	if m.GetByTokenFn != nil {
		return m.GetByTokenFn(ctx, token)
	}
	return nil, nil
}

func (m *MockSessionRepo) GetUserEmail(ctx context.Context, userID string) (string, error) {
	if m.GetUserEmailFn != nil {
		return m.GetUserEmailFn(ctx, userID)
	}
	return "test@example.com", nil
}

// MockPreferenceRepo implements database.PreferenceRepository for testing.
type MockPreferenceRepo struct {
	GetFn    func(ctx context.Context, userID string) (*domain.UserPreference, error)
	UpsertFn func(ctx context.Context, pref *domain.UserPreference) error
}

func (m *MockPreferenceRepo) Get(ctx context.Context, userID string) (*domain.UserPreference, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, userID)
	}
	return &domain.UserPreference{UserID: userID}, nil
}

func (m *MockPreferenceRepo) Upsert(ctx context.Context, pref *domain.UserPreference) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, pref)
	}
	return nil
}

func (m *MockPreferenceRepo) SetTokenWarningSent(ctx context.Context, userID string) error {
	return nil
}

// MockMailer implements mailer.Mailer for testing.
type MockMailer struct {
	SendDigestFn       func(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error
	SendInviteFn       func(ctx context.Context, toEmail, fromName string) error
	SendTokenWarningFn func(ctx context.Context, toEmail string, usagePercent int, tokensUsed, tokenLimit int) error
}

func (m *MockMailer) SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error {
	if m.SendDigestFn != nil {
		return m.SendDigestFn(ctx, toEmail, toName, digest, unsubscribeURL)
	}
	return nil
}

func (m *MockMailer) SendInvite(ctx context.Context, toEmail, fromName string) error {
	if m.SendInviteFn != nil {
		return m.SendInviteFn(ctx, toEmail, fromName)
	}
	return nil
}

func (m *MockMailer) SendTokenWarning(ctx context.Context, toEmail string, usagePercent int, tokensUsed, tokenLimit int) error {
	if m.SendTokenWarningFn != nil {
		return m.SendTokenWarningFn(ctx, toEmail, usagePercent, tokensUsed, tokenLimit)
	}
	return nil
}

// MockSenderPreferenceRepo implements database.SenderPreferenceRepository for testing.
type MockSenderPreferenceRepo struct {
	UpsertFn              func(ctx context.Context, pref *domain.SenderPreference) error
	GetByAddressFn        func(ctx context.Context, userID, fromAddress string) (*domain.SenderPreference, error)
	ListByUserFn          func(ctx context.Context, userID string) ([]domain.SenderPreference, error)
	ListBlockedAddressesFn func(ctx context.Context, userID string) ([]string, error)
}

func (m *MockSenderPreferenceRepo) Upsert(ctx context.Context, pref *domain.SenderPreference) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, pref)
	}
	return nil
}

func (m *MockSenderPreferenceRepo) GetByAddress(ctx context.Context, userID, fromAddress string) (*domain.SenderPreference, error) {
	if m.GetByAddressFn != nil {
		return m.GetByAddressFn(ctx, userID, fromAddress)
	}
	return nil, nil
}

func (m *MockSenderPreferenceRepo) ListByUser(ctx context.Context, userID string) ([]domain.SenderPreference, error) {
	if m.ListByUserFn != nil {
		return m.ListByUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockSenderPreferenceRepo) ListBlockedAddresses(ctx context.Context, userID string) ([]string, error) {
	if m.ListBlockedAddressesFn != nil {
		return m.ListBlockedAddressesFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockSenderPreferenceRepo) MarkUnsubscribed(ctx context.Context, userID, fromAddress string) error {
	return nil
}

// MockInviteRepo implements database.InviteRepository for testing.
type MockInviteRepo struct {
	GetByEmailFn   func(ctx context.Context, email string) (*domain.Invite, error)
	CreateFn       func(ctx context.Context, invite *domain.Invite) error
	ListFn         func(ctx context.Context) ([]domain.Invite, error)
	MarkAcceptedFn func(ctx context.Context, email string) error
}

func (m *MockInviteRepo) GetByEmail(ctx context.Context, email string) (*domain.Invite, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *MockInviteRepo) Create(ctx context.Context, invite *domain.Invite) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, invite)
	}
	return nil
}

func (m *MockInviteRepo) List(ctx context.Context) ([]domain.Invite, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *MockInviteRepo) MarkAccepted(ctx context.Context, email string) error {
	if m.MarkAcceptedFn != nil {
		return m.MarkAcceptedFn(ctx, email)
	}
	return nil
}

// MockProvider implements llm.Provider for testing.
type MockProvider struct {
	TriageEmailFn    func(ctx context.Context, subject, fromAddress, contentPreview string) (*llm.TriageResult, *llm.Usage, error)
	ExtractEmailFn   func(ctx context.Context, subject, content, fromAddress string) (*llm.ExtractionResult, *llm.Usage, error)
	GenerateDigestFn func(ctx context.Context, extractions []domain.Extraction, periodType string, opts *llm.DigestOptions) (*llm.DigestSummary, *llm.Usage, error)
	NameVal          string
	ModelVal         string
}

func (m *MockProvider) TriageEmail(ctx context.Context, subject, fromAddress, contentPreview string) (*llm.TriageResult, *llm.Usage, error) {
	if m.TriageEmailFn != nil {
		return m.TriageEmailFn(ctx, subject, fromAddress, contentPreview)
	}
	return &llm.TriageResult{Extract: true, Reason: "default mock triage"}, &llm.Usage{TotalTokens: 5}, nil
}

func (m *MockProvider) ExtractEmail(ctx context.Context, subject, content, fromAddress string) (*llm.ExtractionResult, *llm.Usage, error) {
	if m.ExtractEmailFn != nil {
		return m.ExtractEmailFn(ctx, subject, content, fromAddress)
	}
	return &llm.ExtractionResult{Summary: "test summary"}, &llm.Usage{TotalTokens: 10}, nil
}

func (m *MockProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string, opts *llm.DigestOptions) (*llm.DigestSummary, *llm.Usage, error) {
	if m.GenerateDigestFn != nil {
		return m.GenerateDigestFn(ctx, extractions, periodType, opts)
	}
	return &llm.DigestSummary{Title: "Test Digest", Summary: "test"}, &llm.Usage{TotalTokens: 10}, nil
}

func (m *MockProvider) Name() string {
	if m.NameVal != "" {
		return m.NameVal
	}
	return "test"
}

func (m *MockProvider) Model() string {
	if m.ModelVal != "" {
		return m.ModelVal
	}
	return "test-model"
}

// MockWaitlistRepo implements database.WaitlistRepository for testing.
type MockWaitlistRepo struct {
	CreateFn     func(ctx context.Context, req *domain.WaitlistRequest) error
	GetByEmailFn func(ctx context.Context, email string) (*domain.WaitlistRequest, error)
	ListFn       func(ctx context.Context) ([]domain.WaitlistRequest, error)
	DeleteFn     func(ctx context.Context, id string) error
}

func (m *MockWaitlistRepo) Create(ctx context.Context, req *domain.WaitlistRequest) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, req)
	}
	return nil
}

func (m *MockWaitlistRepo) GetByEmail(ctx context.Context, email string) (*domain.WaitlistRequest, error) {
	if m.GetByEmailFn != nil {
		return m.GetByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *MockWaitlistRepo) List(ctx context.Context) ([]domain.WaitlistRequest, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *MockWaitlistRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// MockDigestFeedbackRepo implements database.DigestFeedbackRepository for testing.
type MockDigestFeedbackRepo struct {
	UpsertFn       func(ctx context.Context, fb *domain.DigestFeedback) error
	DeleteFn       func(ctx context.Context, userID, digestID string) error
	GetByDigestIDFn func(ctx context.Context, userID, digestID string) (*domain.DigestFeedback, error)
}

func (m *MockDigestFeedbackRepo) Upsert(ctx context.Context, fb *domain.DigestFeedback) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, fb)
	}
	return nil
}

func (m *MockDigestFeedbackRepo) Delete(ctx context.Context, userID, digestID string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, userID, digestID)
	}
	return nil
}

func (m *MockDigestFeedbackRepo) GetByDigestID(ctx context.Context, userID, digestID string) (*domain.DigestFeedback, error) {
	if m.GetByDigestIDFn != nil {
		return m.GetByDigestIDFn(ctx, userID, digestID)
	}
	return nil, nil
}
