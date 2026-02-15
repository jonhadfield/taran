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
	IsRead         bool
	IsStarred      bool
	IsArchived     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	DigestTimezone  string // IANA timezone
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SenderInfo struct {
	FromAddress string `json:"FromAddress"`
	FromName    string `json:"FromName"`
	EmailCount  int    `json:"EmailCount"`
	Status      string `json:"Status"`
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

type ListOptions struct {
	Limit      int
	Offset     int
	Status     *EmailStatus
	IsRead     *bool
	IsStarred  *bool
	IsArchived *bool
	Since      *time.Time
	Before     *time.Time
	Search     *string
}
