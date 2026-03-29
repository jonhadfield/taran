package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookPayloadRepo struct {
	pool *pgxpool.Pool
}

func NewWebhookPayloadRepo(pool *pgxpool.Pool) *WebhookPayloadRepo {
	return &WebhookPayloadRepo{pool: pool}
}

func (r *WebhookPayloadRepo) Create(ctx context.Context, payload *domain.WebhookPayload) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_payload (id, email_id, raw_body, headers, received_at, size_bytes)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		payload.ID, payload.EmailID, payload.RawBody, payload.Headers,
		payload.ReceivedAt, payload.SizeBytes)
	if err != nil {
		return fmt.Errorf("create webhook payload: %w", err)
	}
	return nil
}

func (r *WebhookPayloadRepo) GetByID(ctx context.Context, id string) (*domain.WebhookPayload, error) {
	var p domain.WebhookPayload
	err := r.pool.QueryRow(ctx,
		`SELECT id, email_id, raw_body, headers, received_at, size_bytes
		 FROM webhook_payload WHERE id = $1`, id).
		Scan(&p.ID, &p.EmailID, &p.RawBody, &p.Headers, &p.ReceivedAt, &p.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("get webhook payload: %w", err)
	}
	return &p, nil
}

func (r *WebhookPayloadRepo) GetByEmailID(ctx context.Context, emailID string) (*domain.WebhookPayload, error) {
	var p domain.WebhookPayload
	err := r.pool.QueryRow(ctx,
		`SELECT id, email_id, raw_body, headers, received_at, size_bytes
		 FROM webhook_payload WHERE email_id = $1`, emailID).
		Scan(&p.ID, &p.EmailID, &p.RawBody, &p.Headers, &p.ReceivedAt, &p.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("get webhook payload by email: %w", err)
	}
	return &p, nil
}

func (r *WebhookPayloadRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM webhook_payload WHERE received_at < $1`, before)
	if err != nil {
		return 0, fmt.Errorf("delete old webhook payloads: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *WebhookPayloadRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM webhook_payload`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count webhook payloads: %w", err)
	}
	return count, nil
}
