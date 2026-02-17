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

func (r *FeedbackRepo) Delete(ctx context.Context, userID, emailID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM email_feedback WHERE user_id = $1 AND email_id = $2`,
		userID, emailID,
	)
	if err != nil {
		return fmt.Errorf("delete feedback: %w", err)
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

func (r *FeedbackRepo) GetSenderStats(ctx context.Context, userID string) ([]domain.SenderFeedbackStat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT em.from_address,
		        COUNT(*) FILTER (WHERE f.rating = 'useful'),
		        COUNT(*) FILTER (WHERE f.rating = 'not_useful')
		 FROM email_feedback f
		 JOIN email em ON em.id = f.email_id
		 WHERE f.user_id = $1
		 GROUP BY em.from_address
		 HAVING COUNT(*) >= 2`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get sender feedback stats: %w", err)
	}
	defer rows.Close()

	var stats []domain.SenderFeedbackStat
	for rows.Next() {
		var s domain.SenderFeedbackStat
		if err := rows.Scan(&s.FromAddress, &s.UsefulCount, &s.NotUsefulCount); err != nil {
			return nil, fmt.Errorf("scan sender feedback stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *FeedbackRepo) GetTopicStats(ctx context.Context, userID string) ([]domain.TopicFeedbackStat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT topic,
		        COUNT(*) FILTER (WHERE f.rating = 'useful'),
		        COUNT(*) FILTER (WHERE f.rating = 'not_useful')
		 FROM email_feedback f
		 JOIN extraction e ON e.email_id = f.email_id
		 CROSS JOIN LATERAL jsonb_array_elements_text(e.topics) AS topic
		 WHERE f.user_id = $1
		 GROUP BY topic
		 HAVING COUNT(*) >= 2`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get topic feedback stats: %w", err)
	}
	defer rows.Close()

	var stats []domain.TopicFeedbackStat
	for rows.Next() {
		var s domain.TopicFeedbackStat
		if err := rows.Scan(&s.Topic, &s.UsefulCount, &s.NotUsefulCount); err != nil {
			return nil, fmt.Errorf("scan topic feedback stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
