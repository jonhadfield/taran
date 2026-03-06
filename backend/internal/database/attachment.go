package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttachmentRepo struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepo(pool *pgxpool.Pool) *AttachmentRepo {
	return &AttachmentRepo{pool: pool}
}

func (r *AttachmentRepo) CreateBatch(ctx context.Context, attachments []domain.EmailAttachment) error {
	if len(attachments) == 0 {
		return nil
	}

	for _, a := range attachments {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO email_attachment (id, email_id, filename, content_type, size_bytes, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			a.ID, a.EmailID, a.Filename, a.ContentType, a.SizeBytes, a.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("create attachment: %w", err)
		}
	}
	return nil
}

func (r *AttachmentRepo) ListByEmailID(ctx context.Context, emailID string) ([]domain.EmailAttachment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email_id, filename, content_type, size_bytes, created_at
		 FROM email_attachment WHERE email_id = $1 ORDER BY created_at`, emailID)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	var attachments []domain.EmailAttachment
	for rows.Next() {
		var a domain.EmailAttachment
		if err := rows.Scan(&a.ID, &a.EmailID, &a.Filename, &a.ContentType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		attachments = append(attachments, a)
	}
	return attachments, nil
}
