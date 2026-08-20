package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/crypto"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookPayloadRepo struct {
	pool      *pgxpool.Pool
	encryptor *crypto.Encryptor
}

func NewWebhookPayloadRepo(pool *pgxpool.Pool, encryptor *crypto.Encryptor) *WebhookPayloadRepo {
	return &WebhookPayloadRepo{pool: pool, encryptor: encryptor}
}

func (r *WebhookPayloadRepo) Create(ctx context.Context, payload *domain.WebhookPayload) error {
	// The stored payload is the complete original message. Encrypt it with the
	// same key as the email bodies, otherwise it becomes a plaintext copy of
	// data that is encrypted everywhere else.
	rawBody := payload.RawBody
	encrypted := false
	if r.encryptor != nil && len(rawBody) > 0 {
		enc, err := r.encryptor.EncryptToString(string(rawBody))
		if err != nil {
			return fmt.Errorf("encrypt raw_body: %w", err)
		}
		rawBody = []byte(enc)
		encrypted = true
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_payload (id, email_id, raw_body, headers, received_at, size_bytes, encrypted)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		payload.ID, payload.EmailID, rawBody, payload.Headers,
		payload.ReceivedAt, payload.SizeBytes, encrypted)
	if err != nil {
		return fmt.Errorf("create webhook payload: %w", err)
	}
	return nil
}

// decryptRawBody restores the original message for a row written with
// encryption enabled. Rows predating encryption carry encrypted = false and are
// returned as-is.
func (r *WebhookPayloadRepo) decryptRawBody(p *domain.WebhookPayload, encrypted bool) error {
	if !encrypted || len(p.RawBody) == 0 {
		return nil
	}
	if r.encryptor == nil {
		return fmt.Errorf("payload %s is encrypted but no encryptor is configured", p.ID)
	}
	plain, err := r.encryptor.DecryptFromString(string(p.RawBody))
	if err != nil {
		return fmt.Errorf("decrypt raw_body for payload %s: %w", p.ID, err)
	}
	p.RawBody = []byte(plain)
	return nil
}

func (r *WebhookPayloadRepo) GetByID(ctx context.Context, id string) (*domain.WebhookPayload, error) {
	var p domain.WebhookPayload
	var encrypted bool
	err := r.pool.QueryRow(ctx,
		`SELECT id, email_id, raw_body, headers, received_at, size_bytes, encrypted
		 FROM webhook_payload WHERE id = $1`, id).
		Scan(&p.ID, &p.EmailID, &p.RawBody, &p.Headers, &p.ReceivedAt, &p.SizeBytes, &encrypted)
	if err != nil {
		return nil, fmt.Errorf("get webhook payload: %w", err)
	}
	if err := r.decryptRawBody(&p, encrypted); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *WebhookPayloadRepo) GetByEmailID(ctx context.Context, emailID string) (*domain.WebhookPayload, error) {
	var p domain.WebhookPayload
	var encrypted bool
	err := r.pool.QueryRow(ctx,
		`SELECT id, email_id, raw_body, headers, received_at, size_bytes, encrypted
		 FROM webhook_payload WHERE email_id = $1`, emailID).
		Scan(&p.ID, &p.EmailID, &p.RawBody, &p.Headers, &p.ReceivedAt, &p.SizeBytes, &encrypted)
	if err != nil {
		return nil, fmt.Errorf("get webhook payload by email: %w", err)
	}
	if err := r.decryptRawBody(&p, encrypted); err != nil {
		return nil, err
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
