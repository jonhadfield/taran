package domain

import "time"

// DefaultMonthlyTokenLimit is the fallback monthly token limit when no
// user-specific or admin-configured value exists.
const DefaultMonthlyTokenLimit = 500000

// ValidRatings lists the accepted feedback rating values.
var ValidRatings = map[string]bool{
	"useful":     true,
	"not_useful": true,
}

// IsValidRating returns true if the given rating string is an accepted value.
func IsValidRating(rating string) bool {
	return ValidRatings[rating]
}

// ValidCategories lists the accepted email source categories.
// Used for validation in handlers that accept user-specified categories.
var ValidCategories = map[string]bool{
	"newsletter":    true,
	"personal":      true,
	"transactional": true,
	"marketing":     true,
	"notification":  true,
	"other":         true,
}

type EmailStatus string

const (
	EmailStatusPending    EmailStatus = "pending"
	EmailStatusProcessing EmailStatus = "processing"
	EmailStatusProcessed  EmailStatus = "processed"
	EmailStatusFailed     EmailStatus = "failed"
	EmailStatusSkipped    EmailStatus = "skipped"
)

type Email struct {
	ID             string      `json:"ID"`
	UserID         string      `json:"UserID"`
	AccountID      string      `json:"AccountID"`
	MessageID      string      `json:"MessageID"`
	InReplyTo      string      `json:"InReplyTo"`
	ThreadID       string      `json:"ThreadID"`
	FromAddress    string      `json:"FromAddress"`
	FromName       string      `json:"FromName"`
	ToAddress      string      `json:"ToAddress"`
	Subject        string      `json:"Subject"`
	TextBody       string      `json:"TextBody"`
	HTMLBody       string      `json:"HTMLBody"`
	ReceivedAt     time.Time   `json:"ReceivedAt"`
	DateHeader     time.Time   `json:"DateHeader"`
	Status         EmailStatus `json:"Status"`
	StatusReason   string      `json:"StatusReason"`
	IsRead         bool        `json:"IsRead"`
	IsStarred      bool        `json:"IsStarred"`
	IsArchived        bool      `json:"IsArchived"`
	UnsubscribeURL    string    `json:"UnsubscribeURL"`
	UnsubscribeMailto string    `json:"UnsubscribeMailto"`
	RetryCount        int       `json:"RetryCount"`
	CreatedAt         time.Time `json:"CreatedAt"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
}

type EmailState struct {
	IsRead     *bool `json:"IsRead"`
	IsStarred  *bool `json:"IsStarred"`
	IsArchived *bool `json:"IsArchived"`
}

type Extraction struct {
	ID             string    `json:"ID"`
	EmailID        string    `json:"EmailID"`
	Summary        string    `json:"Summary"`
	KeyPoints      []string  `json:"KeyPoints"`
	Topics         []string  `json:"Topics"`
	Links          []Link    `json:"Links"`
	ActionItems    []string  `json:"ActionItems"`
	Sentiment      string    `json:"Sentiment"`
	SourceCategory string    `json:"SourceCategory"`
	Provider       string    `json:"Provider"`
	Model          string    `json:"Model"`
	TokensUsed     int       `json:"TokensUsed"`
	ProcessedAt    time.Time `json:"ProcessedAt"`
	CreatedAt      time.Time `json:"CreatedAt"`
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Digest struct {
	ID          string      `json:"ID"`
	UserID      string      `json:"UserID"`
	Title       string      `json:"Title"`
	Summary     string      `json:"Summary"`
	Highlights  []string    `json:"Highlights"`
	TopTopics   []string    `json:"TopTopics"`
	PeriodStart time.Time   `json:"PeriodStart"`
	PeriodEnd   time.Time   `json:"PeriodEnd"`
	PeriodType  string      `json:"PeriodType"`
	EmailCount  int         `json:"EmailCount"`
	TokensUsed  int         `json:"TokensUsed"`
	Provider    string      `json:"Provider"`
	Model       string      `json:"Model"`
	GeneratedAt time.Time   `json:"GeneratedAt"`
	SentAt      *time.Time  `json:"SentAt"`
	CreatedAt   time.Time   `json:"CreatedAt"`
	ShareToken  *string     `json:"ShareToken"`
	Items       []DigestItem `json:"Items"`

	// Transient — populated at generation time for email rendering, not persisted
	EmailSummaries []DigestEmailSummary `json:"-"`
}

type DigestEmailSummary struct {
	EmailID     string   `json:"EmailID"`
	Subject     string   `json:"Subject"`
	SenderName  string   `json:"SenderName"`
	Summary     string   `json:"Summary"`
	ActionItems []string `json:"ActionItems"`
	Category    string   `json:"Category"`
}

type WeeklySummary struct {
	ID          string         `json:"ID"`
	UserID      string         `json:"UserID"`
	PeriodStart time.Time      `json:"PeriodStart"`
	PeriodEnd   time.Time      `json:"PeriodEnd"`
	EmailCount  int            `json:"EmailCount"`
	TopSenders  []SenderCount  `json:"TopSenders"`
	Categories  map[string]int `json:"Categories"`
	ActionItems int            `json:"ActionItems"`
	SentAt      *time.Time     `json:"SentAt"`
	CreatedAt   time.Time      `json:"CreatedAt"`
}

type AuditEntry struct {
	ID        string    `json:"ID"`
	UserID    string    `json:"UserID"`
	UserEmail string    `json:"UserEmail"`
	Action    string    `json:"Action"`
	Target    string    `json:"Target,omitempty"`
	Detail    string    `json:"Detail,omitempty"`
	IPAddress string    `json:"IPAddress,omitempty"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type DigestItem struct {
	ID           string `json:"ID"`
	DigestID     string `json:"DigestID"`
	EmailID      string `json:"EmailID"`
	ExtractionID string `json:"ExtractionID"`
	SortOrder    int    `json:"SortOrder"`
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
	ID           string    `json:"ID"`
	UserID       string    `json:"UserID"`
	EmailAddress string    `json:"EmailAddress"`
	DisplayName  string    `json:"DisplayName"`
	IsActive     bool      `json:"IsActive"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
}

type UserPreference struct {
	UserID          string    `json:"UserID"`
	DigestEmail     bool      `json:"DigestEmail"`
	DigestFrequency string    `json:"DigestFrequency"` // "daily" or "weekly"
	DigestHour      int       `json:"DigestHour"`      // 0-23
	DigestDay       int       `json:"DigestDay"`       // 0=Sunday..6=Saturday (used for weekly)
	DigestTimezone  string    `json:"DigestTimezone"`  // IANA timezone
	TopicLimit      int       `json:"TopicLimit"`      // max topics shown in inbox cloud (default 15)
	DigestStyle       string    `json:"DigestStyle"` // "detailed" or "concise"
	InterestKeywords  []string  `json:"InterestKeywords"`
	ExclusionKeywords []string  `json:"ExclusionKeywords"`
	ColorTheme        string    `json:"ColorTheme"`
	MonthlyTokenLimit    int        `json:"MonthlyTokenLimit"`
	ExcludedCategories   []string   `json:"ExcludedCategories"` // categories excluded from digests (default: notification, transactional, marketing)
	TokenWarningSentAt   *time.Time `json:"TokenWarningSentAt"`
	DailyTokenLimit      int        `json:"DailyTokenLimit"` // 0 = no daily limit
	QuietHoursEnabled    bool       `json:"QuietHoursEnabled"`
	QuietHoursStart      int        `json:"QuietHoursStart"` // 0-23, hour in DigestTimezone
	QuietHoursEnd        int        `json:"QuietHoursEnd"`   // 0-23, hour in DigestTimezone
	WeeklySummary        bool       `json:"WeeklySummary"`   // opt-in for weekly activity summary email
	DigestWebhook        bool       `json:"DigestWebhook"`   // enable webhook delivery for digests
	WebhookURL           string     `json:"WebhookURL"`      // URL to POST digest summaries to
	CreatedAt            time.Time  `json:"CreatedAt"`
	UpdatedAt            time.Time  `json:"UpdatedAt"`
}

type Session struct {
	ID        string    `json:"ID"`
	UserID    string    `json:"UserID"`
	UserEmail string    `json:"UserEmail"`
	Token     string    `json:"Token"`
	ExpiresAt time.Time `json:"ExpiresAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

type SenderPreference struct {
	ID             string     `json:"ID"`
	UserID         string     `json:"UserID"`
	FromAddress    string     `json:"FromAddress"`
	Status         string     `json:"Status"`   // "normal", "muted", "blocked", "favorite"
	Category       string     `json:"Category"` // user override: "newsletter", "personal", "transactional", "marketing", "notification", "other", or "" (auto)
	UnsubscribedAt *time.Time `json:"UnsubscribedAt"`
	CreatedAt      time.Time  `json:"CreatedAt"`
	UpdatedAt      time.Time  `json:"UpdatedAt"`
}

type SubscriptionInfo struct {
	FromAddress       string     `json:"FromAddress"`
	FromName          string     `json:"FromName"`
	EmailCount        int        `json:"EmailCount"`
	LastSeen          time.Time  `json:"LastSeen"`
	UnsubscribeURL    string     `json:"UnsubscribeURL"`
	UnsubscribeMailto string     `json:"UnsubscribeMailto"`
	UnsubscribedAt    *time.Time `json:"UnsubscribedAt"`
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
	ID        string    `json:"ID"`
	UserID    string    `json:"UserID"`
	EmailID   string    `json:"EmailID"`
	Rating    string    `json:"Rating"` // "useful" or "not_useful"
	CreatedAt time.Time `json:"CreatedAt"`
}

type SenderFeedbackStat struct {
	FromAddress    string `json:"FromAddress"`
	UsefulCount    int    `json:"UsefulCount"`
	NotUsefulCount int    `json:"NotUsefulCount"`
}

type TopicFeedbackStat struct {
	Topic          string `json:"Topic"`
	UsefulCount    int    `json:"UsefulCount"`
	NotUsefulCount int    `json:"NotUsefulCount"`
}

type ProcessingStats struct {
	PendingCount    int `json:"PendingCount"`
	ProcessingCount int `json:"ProcessingCount"`
	FailedCount     int `json:"FailedCount"`
}

type CategoryCount struct {
	Category string `json:"Category"`
	Count    int    `json:"Count"`
}

type TopicCount struct {
	Topic string `json:"Topic"`
	Count int    `json:"Count"`
}

type HeatmapCell struct {
	Hour int `json:"Hour"` // 0-23
	Day  int `json:"Day"`  // 0=Sunday..6=Saturday
	Count int `json:"Count"`
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
	// Analytics
	ProcessedCount    int              `json:"ProcessedCount"`
	FailedCount       int              `json:"FailedCount"`
	SkippedCount      int              `json:"SkippedCount"`
	PendingCount      int              `json:"PendingCount"`
	FeedbackUseful    int              `json:"FeedbackUseful"`
	FeedbackNotUseful int              `json:"FeedbackNotUseful"`
	WeeklyEmails      []WeekCount      `json:"WeeklyEmails"`
	WeeklyDigests     []WeekCount      `json:"WeeklyDigests"`
	WeeklyTokens      []WeekCount      `json:"WeeklyTokens"`
}

type AdminUser struct {
	ID                string `json:"ID"`
	Email             string `json:"Email"`
	Name              string `json:"Name"`
	EmailCount        int    `json:"EmailCount"`
	MonthlyTokensUsed int    `json:"MonthlyTokensUsed"`
	MonthlyTokenLimit int    `json:"MonthlyTokenLimit"`
}

type WeekCount struct {
	Week  time.Time `json:"Week"`
	Count int       `json:"Count"`
}

type Invite struct {
	ID         string     `json:"ID"`
	Email      string     `json:"Email"`
	InvitedBy  string     `json:"InvitedBy"`
	CreatedAt  time.Time  `json:"CreatedAt"`
	AcceptedAt *time.Time `json:"AcceptedAt"`
}

type WaitlistRequest struct {
	ID        string    `json:"ID"`
	Email     string    `json:"Email"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type DigestFeedback struct {
	ID        string    `json:"ID"`
	UserID    string    `json:"UserID"`
	DigestID  string    `json:"DigestID"`
	Rating    string    `json:"Rating"` // "useful" or "not_useful"
	CreatedAt time.Time `json:"CreatedAt"`
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
	ID           string    `json:"ID"`
	UserID       string    `json:"UserID"`
	Operation    string    `json:"Operation"` // "triage", "extract", "digest"
	Provider     string    `json:"Provider"`
	Model        string    `json:"Model"`
	InputTokens  int       `json:"InputTokens"`
	OutputTokens int       `json:"OutputTokens"`
	TotalTokens  int       `json:"TotalTokens"`
	CreatedAt    time.Time `json:"CreatedAt"`
}

type UsageStats struct {
	MonthlyTokensUsed  int `json:"MonthlyTokensUsed"`
	MonthlyTokenLimit  int `json:"MonthlyTokenLimit"` // 0 = unlimited
	DailyTokensUsed    int `json:"DailyTokensUsed"`
	DailyTokenLimit    int `json:"DailyTokenLimit"` // 0 = no daily limit
	TriageTokens       int `json:"TriageTokens"`
	ExtractTokens      int `json:"ExtractTokens"`
	DigestTokens       int `json:"DigestTokens"`
	DailyHistory       []DailyTokenCount `json:"DailyHistory"`
	PeriodStart        time.Time `json:"PeriodStart"`
	PeriodEnd          time.Time `json:"PeriodEnd"`
}

type DailyTokenCount struct {
	Date   string `json:"Date"`   // "2006-01-02"
	Tokens int    `json:"Tokens"`
}

type UserLLMKey struct {
	ID           string    `json:"ID"`
	UserID       string    `json:"UserID"`
	Provider     string    `json:"Provider"` // "anthropic", "openai"
	EncryptedKey []byte    `json:"EncryptedKey"`
	KeyNonce     []byte    `json:"KeyNonce"`
	KeyHint      string    `json:"KeyHint"` // last 4 chars of the API key
	Model        string    `json:"Model"`   // optional model override
	IsActive     bool      `json:"IsActive"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
}

type WebhookPayload struct {
	ID         string            `json:"ID"`
	EmailID    *string           `json:"EmailID"`
	RawBody    []byte            `json:"RawBody"`
	Headers    map[string]string `json:"Headers"`
	ReceivedAt time.Time         `json:"ReceivedAt"`
	SizeBytes  int               `json:"SizeBytes"`
}

type FailedEmail struct {
	Email
	UserEmail string `json:"UserEmail"`
}

type AutoArchiveRule struct {
	ID               string    `json:"ID"`
	UserID           string    `json:"UserID"`
	RuleType         string    `json:"RuleType"` // "category" or "sender"
	RuleValue        string    `json:"RuleValue"`
	ArchiveAfterDays int       `json:"ArchiveAfterDays"`
	IsActive         bool      `json:"IsActive"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
}

type ArchiveCandidate struct {
	EmailID string `json:"EmailID"`
	UserID  string `json:"UserID"`
}

type Label struct {
	ID         string    `json:"ID"`
	UserID     string    `json:"UserID"`
	Name       string    `json:"Name"`
	Color      string    `json:"Color"`
	EmailCount int       `json:"EmailCount"`
	CreatedAt  time.Time `json:"CreatedAt"`
	UpdatedAt  time.Time `json:"UpdatedAt"`
}

type SavedSearch struct {
	ID        string                 `json:"ID"`
	UserID    string                 `json:"UserID"`
	Name      string                 `json:"Name"`
	Filters   map[string]interface{} `json:"Filters"`
	CreatedAt time.Time              `json:"CreatedAt"`
	UpdatedAt time.Time              `json:"UpdatedAt"`
}

type ListOptions struct {
	Limit         int          `json:"Limit"`
	Offset        int          `json:"Offset"`
	Status        *EmailStatus `json:"Status"`
	IsRead        *bool        `json:"IsRead"`
	IsStarred     *bool        `json:"IsStarred"`
	IsArchived    *bool        `json:"IsArchived"`
	Since         *time.Time   `json:"Since"`
	Before        *time.Time   `json:"Before"`
	Search        *string      `json:"Search"`
	Topic         *string      `json:"Topic"`
	Category      *string      `json:"Category"`
	FromAddress   *string      `json:"FromAddress"`
	HasAttachment *bool        `json:"HasAttachment"`
	LabelID       *string      `json:"LabelID"`
	Sort          string       `json:"Sort"` // "newest" (default), "oldest", "relevance"
}
