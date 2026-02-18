package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferenceRepo struct {
	pool *pgxpool.Pool
}

func NewPreferenceRepo(pool *pgxpool.Pool) *PreferenceRepo {
	return &PreferenceRepo{pool: pool}
}

func (r *PreferenceRepo) Get(ctx context.Context, userID string) (*domain.UserPreference, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, digest_email, digest_frequency, digest_hour, digest_day, digest_timezone, created_at, updated_at
		 FROM user_preference WHERE user_id = $1`, userID)

	var p domain.UserPreference
	err := row.Scan(&p.UserID, &p.DigestEmail, &p.DigestFrequency, &p.DigestHour, &p.DigestDay, &p.DigestTimezone, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return &domain.UserPreference{
			UserID:          userID,
			DigestEmail:     false,
			DigestFrequency: "daily",
			DigestHour:      7,
			DigestDay:       1, // Monday
			DigestTimezone:  "UTC",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get preference: %w", err)
	}
	return &p, nil
}

func (r *PreferenceRepo) Upsert(ctx context.Context, pref *domain.UserPreference) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_preference (user_id, digest_email, digest_frequency, digest_hour, digest_day, digest_timezone, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE SET
		     digest_email = $2, digest_frequency = $3, digest_hour = $4, digest_day = $5, digest_timezone = $6,
		     updated_at = NOW()`,
		pref.UserID, pref.DigestEmail, pref.DigestFrequency, pref.DigestHour, pref.DigestDay, pref.DigestTimezone)
	if err != nil {
		return fmt.Errorf("upsert preference: %w", err)
	}
	return nil
}
