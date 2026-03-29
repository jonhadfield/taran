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

func (r *AttachmentRepo) ListByEmailID(ctx context.Context, emailID string, userID ...string) ([]domain.EmailAttachment, error) {
	// Defence-in-depth: scope to user if provided
	query := `SELECT ea.id, ea.email_id, ea.filename, ea.content_type, ea.size_bytes, ea.created_at
		 FROM email_attachment ea`
	args := []any{emailID}
	if len(userID) > 0 && userID[0] != "" {
		query += ` JOIN email e ON e.id = ea.email_id WHERE ea.email_id = $1 AND e.user_id = $2`
		args = append(args, userID[0])
	} else {
		query += ` WHERE ea.email_id = $1`
	}
	query += ` ORDER BY ea.created_at`
	rows, err := r.pool.Query(ctx, query, args...)
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
