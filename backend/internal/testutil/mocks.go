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
}

// MockEmailRepo implements database.EmailRepository for testing.
type MockEmailRepo struct {
	CreateFn            func(ctx context.Context, email *domain.Email) error
	GetByIDFn           func(ctx context.Context, userID, id string) (*domain.Email, error)
	GetByIDInternalFn   func(ctx context.Context, id string) (*domain.Email, error)
	ListFn              func(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error)
	UpdateStateFn       func(ctx context.Context, userID, id string, state domain.EmailState) error
	GetByMessageIDFn    func(ctx context.Context, messageID string) (*domain.Email, error)
	ListPendingFn       func(ctx context.Context, limit int) ([]domain.Email, error)
	SetStatusFn         func(ctx context.Context, id string, status domain.EmailStatus) error
	ListActiveUserIDsFn func(ctx context.Context, from, to time.Time) ([]string, error)

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

func (m *MockEmailRepo) SetStatus(ctx context.Context, id string, status domain.EmailStatus) error {
	m.mu.Lock()
	m.SetStatusCalls = append(m.SetStatusCalls, SetStatusCall{ID: id, Status: status})
	m.mu.Unlock()
	if m.SetStatusFn != nil {
		return m.SetStatusFn(ctx, id, status)
	}
	return nil
}

func (m *MockEmailRepo) ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error) {
	if m.ListActiveUserIDsFn != nil {
		return m.ListActiveUserIDsFn(ctx, from, to)
	}
	return nil, nil
}

// MockExtractionRepo implements database.ExtractionRepository for testing.
type MockExtractionRepo struct {
	CreateFn              func(ctx context.Context, extraction *domain.Extraction) error
	GetByEmailIDFn        func(ctx context.Context, emailID string) (*domain.Extraction, error)
	ListByUserAndPeriodFn func(ctx context.Context, userID string, from, to time.Time) ([]domain.Extraction, error)
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

func (m *MockExtractionRepo) ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time) ([]domain.Extraction, error) {
	if m.ListByUserAndPeriodFn != nil {
		return m.ListByUserAndPeriodFn(ctx, userID, from, to)
	}
	return nil, nil
}

// MockDigestRepo implements database.DigestRepository for testing.
type MockDigestRepo struct {
	CreateFn          func(ctx context.Context, digest *domain.Digest) error
	GetByIDFn         func(ctx context.Context, userID, id string) (*domain.Digest, error)
	ListFn            func(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error)
	SetSentAtFn       func(ctx context.Context, id string, sentAt time.Time) error
	SetShareTokenFn   func(ctx context.Context, id, userID, token string) error
	ClearShareTokenFn func(ctx context.Context, id, userID string) error
	GetByShareTokenFn func(ctx context.Context, token string) (*domain.Digest, error)
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

func (m *MockDigestRepo) List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID, opts)
	}
	return nil, 0, nil
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

// MockMailer implements mailer.Mailer for testing.
type MockMailer struct {
	SendDigestFn func(ctx context.Context, toEmail, toName string, digest *domain.Digest) error
}

func (m *MockMailer) SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest) error {
	if m.SendDigestFn != nil {
		return m.SendDigestFn(ctx, toEmail, toName, digest)
	}
	return nil
}

// MockProvider implements llm.Provider for testing.
type MockProvider struct {
	TriageEmailFn    func(ctx context.Context, subject, fromAddress, contentPreview string) (*llm.TriageResult, *llm.Usage, error)
	ExtractEmailFn   func(ctx context.Context, subject, content, fromAddress string) (*llm.ExtractionResult, *llm.Usage, error)
	GenerateDigestFn func(ctx context.Context, extractions []domain.Extraction, periodType string) (*llm.DigestSummary, *llm.Usage, error)
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

func (m *MockProvider) GenerateDigest(ctx context.Context, extractions []domain.Extraction, periodType string) (*llm.DigestSummary, *llm.Usage, error) {
	if m.GenerateDigestFn != nil {
		return m.GenerateDigestFn(ctx, extractions, periodType)
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
