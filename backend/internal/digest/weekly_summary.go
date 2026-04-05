package digest

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

// WeeklySummaryRunner generates weekly activity summaries.
type WeeklySummaryRunner struct {
	mu          sync.Mutex
	Emails      database.EmailRepository
	Extractions database.ExtractionRepository
	Preferences database.PreferenceRepository
	Summaries   *database.WeeklySummaryRepo
}

// Run generates weekly summaries for all eligible users.
// Called from the same hourly cron as digests — only sends on Sundays at each user's preferred hour.
// Uses TryLock to ensure only one run executes at a time, preventing duplicate
// summaries from concurrent cron triggers.
func (r *WeeklySummaryRunner) Run() {
	if !r.mu.TryLock() {
		slog.Info("weekly summary generation already in progress, skipping")
		return
	}
	defer r.mu.Unlock()

	ctx := context.Background()
	nowUTC := time.Now().UTC()

	// Find users who had emails in the last 7 days
	periodEnd := nowUTC
	periodStart := nowUTC.Add(-7 * 24 * time.Hour)

	// List all user IDs with emails in this period
	userIDs, err := r.Emails.ListActiveUserIDs(ctx, periodStart, periodEnd)
	if err != nil {
		slog.Error("weekly summary: failed to list active users", "error", err)
		return
	}

	generated := 0

	for _, userID := range userIDs {
		pref, err := r.Preferences.Get(ctx, userID)
		if err != nil {
			continue
		}

		if !pref.WeeklySummary {
			continue
		}

		if !shouldSendWeeklySummary(pref, nowUTC) {
			continue
		}

		// Check if a summary was already generated for this user and period
		exists, err := r.Summaries.ExistsForPeriod(ctx, userID, periodStart, periodEnd)
		if err != nil {
			slog.Error("weekly summary: failed to check existing", "userID", userID, "error", err)
			continue
		}
		if exists {
			continue
		}

		summary, err := r.GenerateForUser(ctx, userID, periodStart, periodEnd)
		if err != nil {
			slog.Error("weekly summary: generation failed", "userID", userID, "error", err)
			continue
		}
		if summary == nil {
			continue // No activity
		}

		if err := r.Summaries.Create(ctx, summary); err != nil {
			slog.Error("weekly summary: failed to store", "userID", userID, "error", err)
			continue
		}
		generated++
	}

	if generated > 0 {
		slog.Info("weekly summaries complete", "users", len(userIDs), "generated", generated)
	}
}

// GenerateForUser creates a weekly summary for a single user over the given period.
// Returns nil if the user had no email activity in the period.
func (r *WeeklySummaryRunner) GenerateForUser(ctx context.Context, userID string, periodStart, periodEnd time.Time) (*domain.WeeklySummary, error) {
	emailCount, err := r.Emails.CountByPeriod(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	if emailCount == 0 {
		return nil, nil
	}

	topSenders, _ := r.Emails.TopSenders(ctx, userID, periodStart, periodEnd, 5)
	actionItems, _ := r.Extractions.CountActionItems(ctx, userID, periodStart, periodEnd)

	// Category breakdown from extractions in period
	extractions, _ := r.Extractions.ListByUserAndPeriod(ctx, userID, periodStart, periodEnd)
	categories := make(map[string]int)
	for _, ext := range extractions {
		cat := ext.SourceCategory
		if cat == "" {
			cat = "other"
		}
		categories[cat]++
	}

	return &domain.WeeklySummary{
		ID:          uuid.New().String(),
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		EmailCount:  emailCount,
		TopSenders:  topSenders,
		Categories:  categories,
		ActionItems: actionItems,
		CreatedAt:   time.Now(),
	}, nil
}

func shouldSendWeeklySummary(pref *domain.UserPreference, nowUTC time.Time) bool {
	loc, err := time.LoadLocation(pref.DigestTimezone)
	if err != nil {
		loc = time.UTC
	}

	userNow := nowUTC.In(loc)

	// Send on Sundays at the user's preferred digest hour
	if userNow.Weekday() != time.Sunday {
		return false
	}

	return userNow.Hour() == pref.DigestHour
}
