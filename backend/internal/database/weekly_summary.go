package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WeeklySummaryRepo struct {
	pool *pgxpool.Pool
}

func NewWeeklySummaryRepo(pool *pgxpool.Pool) *WeeklySummaryRepo {
	return &WeeklySummaryRepo{pool: pool}
}

func (r *WeeklySummaryRepo) Create(ctx context.Context, s *domain.WeeklySummary) error {
	sendersJSON, _ := json.Marshal(s.TopSenders)
	categoriesJSON, _ := json.Marshal(s.Categories)

	_, err := r.pool.Exec(ctx,
		`INSERT INTO weekly_summary (id, user_id, period_start, period_end, email_count, top_senders, categories, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		s.ID, s.UserID, s.PeriodStart, s.PeriodEnd, s.EmailCount,
		sendersJSON, categoriesJSON, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create weekly summary: %w", err)
	}
	return nil
}

func (r *WeeklySummaryRepo) ExistsForPeriod(ctx context.Context, userID string, periodStart, periodEnd time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM weekly_summary WHERE user_id = $1 AND period_start = $2 AND period_end = $3)`,
		userID, periodStart, periodEnd,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check weekly summary exists: %w", err)
	}
	return exists, nil
}

func (r *WeeklySummaryRepo) SetSentAt(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE weekly_summary SET sent_at = $1 WHERE id = $2`, sentAt, id)
	if err != nil {
		return fmt.Errorf("set weekly summary sent_at: %w", err)
	}
	return nil
}

func (r *WeeklySummaryRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.WeeklySummary, int, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM weekly_summary WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count weekly summaries: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, period_start, period_end, email_count, top_senders, categories, sent_at, created_at
		 FROM weekly_summary WHERE user_id = $1 ORDER BY period_start DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list weekly summaries: %w", err)
	}
	defer rows.Close()

	var summaries []domain.WeeklySummary
	for rows.Next() {
		var s domain.WeeklySummary
		var topSenders, categories []byte
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PeriodStart, &s.PeriodEnd, &s.EmailCount,
			&topSenders, &categories, &s.SentAt, &s.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan weekly summary: %w", err)
		}
		json.Unmarshal(topSenders, &s.TopSenders)
		json.Unmarshal(categories, &s.Categories)
		summaries = append(summaries, s)
	}
	return summaries, int(total), nil
}

func (r *WeeklySummaryRepo) GetByID(ctx context.Context, userID, id string) (*domain.WeeklySummary, error) {
	var s domain.WeeklySummary
	var topSenders, categories []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, period_start, period_end, email_count, top_senders, categories, sent_at, created_at
		 FROM weekly_summary WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&s.ID, &s.UserID, &s.PeriodStart, &s.PeriodEnd, &s.EmailCount,
		&topSenders, &categories, &s.SentAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get weekly summary: %w", err)
	}
	json.Unmarshal(topSenders, &s.TopSenders)
	json.Unmarshal(categories, &s.Categories)
	return &s, nil
}
