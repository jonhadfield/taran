package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedbackRepo struct {
	pool *pgxpool.Pool
}

func NewFeedbackRepo(pool *pgxpool.Pool) *FeedbackRepo {
	return &FeedbackRepo{pool: pool}
}

func (r *FeedbackRepo) Upsert(ctx context.Context, fb *domain.EmailFeedback) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_feedback (id, user_id, email_id, rating, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, email_id) DO UPDATE SET rating = EXCLUDED.rating`,
		fb.ID, fb.UserID, fb.EmailID, fb.Rating, fb.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert feedback: %w", err)
	}
	return nil
}

func (r *FeedbackRepo) GetByEmailID(ctx context.Context, userID, emailID string) (*domain.EmailFeedback, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_id, rating, created_at
		 FROM email_feedback WHERE user_id = $1 AND email_id = $2`,
		userID, emailID)

	var fb domain.EmailFeedback
	err := row.Scan(&fb.ID, &fb.UserID, &fb.EmailID, &fb.Rating, &fb.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get feedback: %w", err)
	}
	return &fb, nil
}
