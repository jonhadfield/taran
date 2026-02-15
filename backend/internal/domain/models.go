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
	Items       []DigestItem
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
	UserID      string
	DigestEmail bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Session struct {
	ID        string
	UserID    string
	UserEmail string
	Token     string
	ExpiresAt time.Time
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
}
