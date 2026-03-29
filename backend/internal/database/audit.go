package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

// Log records an audit entry.
func (r *AuditRepo) Log(ctx context.Context, userID, userEmail, action, target, detail, ip string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_log (id, user_id, user_email, action, target, detail, ip_address, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		uuid.New().String(), userID, userEmail, action, target, detail, ip,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// List returns recent audit entries, most recent first.
func (r *AuditRepo) List(ctx context.Context, limit int) ([]domain.AuditEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, user_email, action, target, detail, ip_address, created_at
		 FROM audit_log ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()

	var entries []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.UserEmail, &e.Action, &e.Target, &e.Detail, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
