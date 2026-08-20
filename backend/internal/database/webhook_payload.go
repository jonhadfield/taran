package database

import (
	"context"
	"fmt"
	"log/slog"
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

// BackfillStats reports the outcome of an encryption backfill run.
type BackfillStats struct {
	Scanned      int // rows examined
	Encrypted    int // rows rewritten as ciphertext (or that would be, in a dry run)
	SkippedEmpty int // rows with no body to encrypt
	Failed       int // rows left untouched because encrypt/verify failed
}

// BackfillEncryption encrypts webhook payloads written before encryption was
// enabled, leaving already-encrypted rows alone.
//
// Rows are walked by keyset pagination on the primary key rather than by
// repeatedly selecting `encrypted = FALSE`, so a row that fails to encrypt
// cannot stall the scan by reappearing at the head of every batch. Each row is
// verified by decrypting the ciphertext and comparing it to the original before
// the update is issued: a payload rewritten with something that will not
// decrypt is unrecoverable, so it is better to leave it in plaintext and report
// it. The update is additionally guarded on `encrypted = FALSE` so a concurrent
// run cannot double-encrypt a row.
func (r *WebhookPayloadRepo) BackfillEncryption(
	ctx context.Context,
	batchSize int,
	dryRun bool,
	onProgress func(BackfillStats),
) (BackfillStats, error) {
	var stats BackfillStats

	if r.encryptor == nil {
		return stats, fmt.Errorf("no encryptor configured: set TARAN_ENCRYPTION_KEY")
	}
	if batchSize <= 0 {
		batchSize = 100
	}

	lastID := ""
	for {
		rows, err := r.pool.Query(ctx,
			`SELECT id, raw_body
			   FROM webhook_payload
			  WHERE encrypted = FALSE AND id > $1
			  ORDER BY id
			  LIMIT $2`, lastID, batchSize)
		if err != nil {
			return stats, fmt.Errorf("select unencrypted payloads: %w", err)
		}

		type row struct {
			id      string
			rawBody []byte
		}
		var batch []row
		for rows.Next() {
			var rw row
			if err := rows.Scan(&rw.id, &rw.rawBody); err != nil {
				rows.Close()
				return stats, fmt.Errorf("scan payload: %w", err)
			}
			batch = append(batch, rw)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return stats, fmt.Errorf("iterate payloads: %w", err)
		}
		if len(batch) == 0 {
			break
		}

		for _, rw := range batch {
			stats.Scanned++
			lastID = rw.id

			if len(rw.rawBody) == 0 {
				stats.SkippedEmpty++
				continue
			}

			ciphertext, err := r.encryptor.EncryptToString(string(rw.rawBody))
			if err != nil {
				slog.Error("backfill: failed to encrypt payload", "payloadID", rw.id, "error", err)
				stats.Failed++
				continue
			}

			// Prove the row will still be readable before overwriting it.
			plain, err := r.encryptor.DecryptFromString(ciphertext)
			if err != nil || plain != string(rw.rawBody) {
				slog.Error("backfill: verification failed, leaving row untouched",
					"payloadID", rw.id, "error", err)
				stats.Failed++
				continue
			}

			if dryRun {
				stats.Encrypted++
				continue
			}

			tag, err := r.pool.Exec(ctx,
				`UPDATE webhook_payload
				    SET raw_body = $1, encrypted = TRUE
				  WHERE id = $2 AND encrypted = FALSE`, []byte(ciphertext), rw.id)
			if err != nil {
				slog.Error("backfill: failed to update payload", "payloadID", rw.id, "error", err)
				stats.Failed++
				continue
			}
			if tag.RowsAffected() == 0 {
				// Encrypted by a concurrent run between select and update.
				continue
			}
			stats.Encrypted++
		}

		if onProgress != nil {
			onProgress(stats)
		}
		if ctx.Err() != nil {
			return stats, ctx.Err()
		}
	}

	return stats, nil
}

// CountUnencrypted reports how many payloads still hold a plaintext body.
func (r *WebhookPayloadRepo) CountUnencrypted(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM webhook_payload WHERE encrypted = FALSE AND octet_length(raw_body) > 0`).
		Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unencrypted payloads: %w", err)
	}
	return count, nil
}
