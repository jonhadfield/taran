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
		`INSERT INTO weekly_summary (id, user_id, period_start, period_end, email_count, top_senders, categories, action_items, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		s.ID, s.UserID, s.PeriodStart, s.PeriodEnd, s.EmailCount,
		sendersJSON, categoriesJSON, s.ActionItems, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create weekly summary: %w", err)
	}
	return nil
}

func (r *WeeklySummaryRepo) SetSentAt(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE weekly_summary SET sent_at = $1 WHERE id = $2`, sentAt, id)
	if err != nil {
		return fmt.Errorf("set weekly summary sent_at: %w", err)
	}
	return nil
}
