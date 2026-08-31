package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferenceRepo struct {
	pool        *pgxpool.Pool
	AppSettings *AppSettingRepo
}

func NewPreferenceRepo(pool *pgxpool.Pool) *PreferenceRepo {
	return &PreferenceRepo{pool: pool}
}

func (r *PreferenceRepo) defaultTokenLimit(ctx context.Context) int {
	if r.AppSettings != nil {
		limit, err := r.AppSettings.GetInt(ctx, "default_monthly_token_limit", domain.DefaultMonthlyTokenLimit)
		if err == nil {
			return limit
		}
	}
	return domain.DefaultMonthlyTokenLimit
}

func (r *PreferenceRepo) Get(ctx context.Context, userID string) (*domain.UserPreference, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, digest_email, digest_frequency, digest_hour, digest_day, digest_timezone, topic_limit, digest_style, interest_keywords, exclusion_keywords, color_theme, monthly_token_limit, excluded_categories, token_warning_sent_at, quiet_hours_enabled, quiet_hours_start, quiet_hours_end, daily_token_limit, weekly_summary, digest_webhook, webhook_url, created_at, updated_at
		 FROM user_preference WHERE user_id = $1`, userID)

	var p domain.UserPreference
	var interestRaw, exclusionRaw, excludedCatsRaw []byte
	err := row.Scan(&p.UserID, &p.DigestEmail, &p.DigestFrequency, &p.DigestHour, &p.DigestDay, &p.DigestTimezone, &p.TopicLimit, &p.DigestStyle, &interestRaw, &exclusionRaw, &p.ColorTheme, &p.MonthlyTokenLimit, &excludedCatsRaw, &p.TokenWarningSentAt, &p.QuietHoursEnabled, &p.QuietHoursStart, &p.QuietHoursEnd, &p.DailyTokenLimit, &p.WeeklySummary, &p.DigestWebhook, &p.WebhookURL, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return &domain.UserPreference{
			UserID:             userID,
			DigestEmail:        false,
			DigestFrequency:    "daily",
			DigestHour:         7,
			DigestDay:          1, // Monday
			DigestTimezone:     "UTC",
			TopicLimit:         15,
			DigestStyle:        "detailed",
			InterestKeywords:   []string{},
			ExclusionKeywords:  []string{},
			ColorTheme:         "brand",
			MonthlyTokenLimit:  r.defaultTokenLimit(ctx),
			ExcludedCategories: []string{"notification", "transactional", "marketing"},
			QuietHoursEnabled:  false,
			QuietHoursStart:    22,
			QuietHoursEnd:      7,
			WeeklySummary:      true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get preference: %w", err)
	}
	p.InterestKeywords = []string{}
	p.ExclusionKeywords = []string{}
	p.ExcludedCategories = []string{"notification", "transactional", "marketing"}
	if len(interestRaw) > 0 {
		_ = json.Unmarshal(interestRaw, &p.InterestKeywords)
	}
	if len(exclusionRaw) > 0 {
		_ = json.Unmarshal(exclusionRaw, &p.ExclusionKeywords)
	}
	if len(excludedCatsRaw) > 0 {
		_ = json.Unmarshal(excludedCatsRaw, &p.ExcludedCategories)
	}
	// If user has no custom limit set, use the global default
	if p.MonthlyTokenLimit == 0 {
		p.MonthlyTokenLimit = r.defaultTokenLimit(ctx)
	}
	return &p, nil
}

func (r *PreferenceRepo) Upsert(ctx context.Context, pref *domain.UserPreference) error {
	if _, err := time.LoadLocation(pref.DigestTimezone); err != nil {
		return fmt.Errorf("invalid digest timezone %q: %w", pref.DigestTimezone, err)
	}

	interestJSON, err := json.Marshal(pref.InterestKeywords)
	if err != nil {
		return fmt.Errorf("marshal interest keywords: %w", err)
	}
	exclusionJSON, err := json.Marshal(pref.ExclusionKeywords)
	if err != nil {
		return fmt.Errorf("marshal exclusion keywords: %w", err)
	}
	excludedCatsJSON, err := json.Marshal(pref.ExcludedCategories)
	if err != nil {
		return fmt.Errorf("marshal excluded categories: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO user_preference (user_id, digest_email, digest_frequency, digest_hour, digest_day, digest_timezone, topic_limit, digest_style, interest_keywords, exclusion_keywords, color_theme, monthly_token_limit, excluded_categories, quiet_hours_enabled, quiet_hours_start, quiet_hours_end, daily_token_limit, weekly_summary, digest_webhook, webhook_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE SET
		     digest_email = $2, digest_frequency = $3, digest_hour = $4, digest_day = $5, digest_timezone = $6, topic_limit = $7, digest_style = $8,
		     interest_keywords = $9, exclusion_keywords = $10, color_theme = $11, monthly_token_limit = $12, excluded_categories = $13,
		     quiet_hours_enabled = $14, quiet_hours_start = $15, quiet_hours_end = $16, daily_token_limit = $17, weekly_summary = $18,
		     digest_webhook = $19, webhook_url = $20,
		     updated_at = NOW()`,
		pref.UserID, pref.DigestEmail, pref.DigestFrequency, pref.DigestHour, pref.DigestDay, pref.DigestTimezone, pref.TopicLimit, pref.DigestStyle, interestJSON, exclusionJSON, pref.ColorTheme, pref.MonthlyTokenLimit, excludedCatsJSON, pref.QuietHoursEnabled, pref.QuietHoursStart, pref.QuietHoursEnd, pref.DailyTokenLimit, pref.WeeklySummary, pref.DigestWebhook, pref.WebhookURL)
	if err != nil {
		return fmt.Errorf("upsert preference: %w", err)
	}
	return nil
}

func (r *PreferenceRepo) SetTokenWarningSent(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_preference SET token_warning_sent_at = NOW() WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("set token warning sent: %w", err)
	}
	return nil
}
