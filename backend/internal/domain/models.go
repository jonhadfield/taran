package domain

import "time"

type EmailStatus string

const (
	EmailStatusPending    EmailStatus = "pending"
	EmailStatusProcessing EmailStatus = "processing"
	EmailStatusProcessed  EmailStatus = "processed"
	EmailStatusFailed     EmailStatus = "failed"
	EmailStatusSkipped    EmailStatus = "skipped"
)

type Email struct {
	ID             string
	UserID         string
	AccountID      string
	MessageID      string
	FromAddress    string
	FromName       string
	ToAddress      string
	Subject        string
	TextBody       string
	HTMLBody       string
	ReceivedAt     time.Time
	DateHeader     time.Time
	Status         EmailStatus
	StatusReason   string
	IsRead         bool
	IsStarred      bool
	IsArchived        bool
	UnsubscribeURL    string
	UnsubscribeMailto string
	RetryCount        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type EmailState struct {
	IsRead     *bool
	IsStarred  *bool
	IsArchived *bool
}

type Extraction struct {
	ID             string
	EmailID        string
	Summary        string
	KeyPoints      []string
	Topics         []string
	Links          []Link
	ActionItems    []string
	Sentiment      string
	SourceCategory string
	Provider       string
	Model          string
	TokensUsed     int
	ProcessedAt    time.Time
	CreatedAt      time.Time
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Digest struct {
	ID          string
	UserID      string
	Title       string
	Summary     string
	Highlights  []string
	TopTopics   []string
	PeriodStart time.Time
	PeriodEnd   time.Time
	PeriodType  string
	EmailCount  int
	TokensUsed  int
	Provider    string
	Model       string
	GeneratedAt time.Time
	SentAt      *time.Time
	CreatedAt   time.Time
	ShareToken  *string
	Items       []DigestItem

	// Transient — populated at generation time for email rendering, not persisted
	EmailSummaries []DigestEmailSummary `json:"-"`
}

type DigestEmailSummary struct {
	EmailID    string
	Subject    string
	SenderName string
	Summary    string
}

type DigestItem struct {
	ID           string
	DigestID     string
	EmailID      string
	ExtractionID string
	SortOrder    int
	// Enriched via JOIN — not stored in digest_item table
	Subject     string `json:"Subject,omitempty"`
	FromName    string `json:"FromName,omitempty"`
	FromAddress string `json:"FromAddress,omitempty"`
	Summary     string `json:"Summary,omitempty"`
}

type EmailAttachment struct {
	ID          string    `json:"ID"`
	EmailID     string    `json:"EmailID"`
	Filename    string    `json:"Filename"`
	ContentType string    `json:"ContentType"`
	SizeBytes   int       `json:"SizeBytes"`
	CreatedAt   time.Time `json:"CreatedAt"`
}

type EmailAccount struct {
	ID           string
	UserID       string
	EmailAddress string
	DisplayName  string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserPreference struct {
	UserID          string
	DigestEmail     bool
	DigestFrequency string // "daily" or "weekly"
	DigestHour      int    // 0-23
	DigestDay       int    // 0=Sunday..6=Saturday (used for weekly)
	DigestTimezone  string // IANA timezone
	TopicLimit      int    // max topics shown in inbox cloud (default 15)
	DigestStyle       string // "detailed" or "concise"
	InterestKeywords  []string
	ExclusionKeywords []string
	ColorTheme        string
	MonthlyTokenLimit    int
	ExcludedCategories   []string // categories excluded from digests (default: notification, transactional, marketing)
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Session struct {
	ID        string
	UserID    string
	UserEmail string
	Token     string
	ExpiresAt time.Time
}

type SenderPreference struct {
	ID          string
	UserID      string
	FromAddress string
	Status      string // "normal", "muted", "blocked", "favorite"
	Category    string // user override: "newsletter", "personal", "transactional", "marketing", "notification", "other", or "" (auto)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SenderInfo struct {
	FromAddress  string `json:"FromAddress"`
	FromName     string `json:"FromName"`
	EmailCount   int    `json:"EmailCount"`
	Status       string `json:"Status"`
	Category     string `json:"Category"`     // user override or auto-detected
	AutoCategory string `json:"AutoCategory"` // most common LLM-detected category
}

type SenderDetail struct {
	FromAddress  string    `json:"FromAddress"`
	FromName     string    `json:"FromName"`
	EmailCount   int       `json:"EmailCount"`
	Status       string    `json:"Status"`
	Category     string    `json:"Category"`
	AutoCategory string    `json:"AutoCategory"`
	FirstSeen    time.Time `json:"FirstSeen"`
	LastSeen     time.Time `json:"LastSeen"`
}

type SenderCount struct {
	FromAddress string `json:"FromAddress"`
	FromName    string `json:"FromName"`
	Count       int    `json:"Count"`
}

type UserStats struct {
	EmailsThisWeek int           `json:"EmailsThisWeek"`
	EmailsLastWeek int           `json:"EmailsLastWeek"`
	TotalEmails    int           `json:"TotalEmails"`
	TopSenders     []SenderCount `json:"TopSenders"`
}

type EmailFeedback struct {
	ID        string
	UserID    string
	EmailID   string
	Rating    string // "useful" or "not_useful"
	CreatedAt time.Time
}

type SenderFeedbackStat struct {
	FromAddress    string
	UsefulCount    int
	NotUsefulCount int
}

type TopicFeedbackStat struct {
	Topic          string
	UsefulCount    int
	NotUsefulCount int
}

type ProcessingStats struct {
	PendingCount    int `json:"PendingCount"`
	ProcessingCount int `json:"ProcessingCount"`
	FailedCount     int `json:"FailedCount"`
}

type AdminStats struct {
	TotalUsers          int           `json:"TotalUsers"`
	ActiveUsersWeek     int           `json:"ActiveUsersWeek"`
	TotalEmails         int           `json:"TotalEmails"`
	EmailsThisWeek      int           `json:"EmailsThisWeek"`
	TotalDigests        int           `json:"TotalDigests"`
	DigestsThisWeek     int           `json:"DigestsThisWeek"`
	TopGlobalSenders    []SenderCount `json:"TopGlobalSenders"`
	LLMProvider         string        `json:"LLMProvider"`
	LLMModel            string        `json:"LLMModel"`
	MonthlyTokensUsed        int           `json:"MonthlyTokensUsed"`
	DefaultMonthlyTokenLimit int           `json:"DefaultMonthlyTokenLimit"`
}

type WeekCount struct {
	Week  time.Time `json:"Week"`
	Count int       `json:"Count"`
}

type Invite struct {
	ID         string
	Email      string
	InvitedBy  string
	CreatedAt  time.Time
	AcceptedAt *time.Time
}

type WaitlistRequest struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type DigestFeedback struct {
	ID        string
	UserID    string
	DigestID  string
	Rating    string // "useful" or "not_useful"
	CreatedAt time.Time
}

type DigestPreviewItem struct {
	EmailID     string `json:"EmailID"`
	Subject     string `json:"Subject"`
	FromName    string `json:"FromName"`
	FromAddress string `json:"FromAddress"`
	Summary     string `json:"Summary"`
	ReceivedAt  time.Time `json:"ReceivedAt"`
}

type DigestPreview struct {
	PeriodStart time.Time          `json:"PeriodStart"`
	PeriodEnd   time.Time          `json:"PeriodEnd"`
	PeriodType  string             `json:"PeriodType"`
	EmailCount  int                `json:"EmailCount"`
	Items       []DigestPreviewItem `json:"Items"`
}

type TokenUsage struct {
	ID           string
	UserID       string
	Operation    string // "triage", "extract", "digest"
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	CreatedAt    time.Time
}

type UsageStats struct {
	MonthlyTokensUsed  int `json:"MonthlyTokensUsed"`
	MonthlyTokenLimit  int `json:"MonthlyTokenLimit"` // 0 = unlimited
	DailyTokensUsed    int `json:"DailyTokensUsed"`
	TriageTokens       int `json:"TriageTokens"`
	ExtractTokens      int `json:"ExtractTokens"`
	DigestTokens       int `json:"DigestTokens"`
	PeriodStart        time.Time `json:"PeriodStart"`
	PeriodEnd          time.Time `json:"PeriodEnd"`
}

type ListOptions struct {
	Limit         int
	Offset        int
	Status        *EmailStatus
	IsRead        *bool
	IsStarred     *bool
	IsArchived    *bool
	Since         *time.Time
	Before        *time.Time
	Search        *string
	Topic         *string
	Category      *string
	FromAddress   *string
	HasAttachment *bool
}
