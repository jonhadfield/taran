package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DigestFeedbackRepo struct {
	pool *pgxpool.Pool
}

func NewDigestFeedbackRepo(pool *pgxpool.Pool) *DigestFeedbackRepo {
	return &DigestFeedbackRepo{pool: pool}
}

func (r *DigestFeedbackRepo) Upsert(ctx context.Context, fb *domain.DigestFeedback) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO digest_feedback (id, user_id, digest_id, rating, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, digest_id) DO UPDATE SET rating = EXCLUDED.rating`,
		fb.ID, fb.UserID, fb.DigestID, fb.Rating, fb.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert digest feedback: %w", err)
	}
	return nil
}

func (r *DigestFeedbackRepo) Delete(ctx context.Context, userID, digestID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM digest_feedback WHERE user_id = $1 AND digest_id = $2`,
		userID, digestID,
	)
	if err != nil {
		return fmt.Errorf("delete digest feedback: %w", err)
	}
	return nil
}

func (r *DigestFeedbackRepo) GetByDigestID(ctx context.Context, userID, digestID string) (*domain.DigestFeedback, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, digest_id, rating, created_at
		 FROM digest_feedback WHERE user_id = $1 AND digest_id = $2`,
		userID, digestID)

	var fb domain.DigestFeedback
	err := row.Scan(&fb.ID, &fb.UserID, &fb.DigestID, &fb.Rating, &fb.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get digest feedback: %w", err)
	}
	return &fb, nil
}
