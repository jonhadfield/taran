package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LabelRepo struct {
	pool *pgxpool.Pool
}

func NewLabelRepo(pool *pgxpool.Pool) *LabelRepo {
	return &LabelRepo{pool: pool}
}

func (r *LabelRepo) Create(ctx context.Context, label *domain.Label) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO label (id, user_id, name, color, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		label.ID, label.UserID, label.Name, label.Color)
	if err != nil {
		return fmt.Errorf("create label: %w", err)
	}
	return nil
}

func (r *LabelRepo) GetByID(ctx context.Context, userID, id string) (*domain.Label, error) {
	var label domain.Label
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, color, created_at, updated_at
		 FROM label WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&label.ID, &label.UserID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get label: %w", err)
	}
	return &label, nil
}

func (r *LabelRepo) ListByUser(ctx context.Context, userID string) ([]domain.Label, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT l.id, l.user_id, l.name, l.color, l.created_at, l.updated_at,
		        (SELECT COUNT(*) FROM email_label el WHERE el.label_id = l.id) AS email_count
		 FROM label l WHERE l.user_id = $1 ORDER BY l.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var label domain.Label
		if err := rows.Scan(&label.ID, &label.UserID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt, &label.EmailCount); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func (r *LabelRepo) Update(ctx context.Context, userID string, label *domain.Label) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE label SET name = $3, color = $4, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2`,
		label.ID, userID, label.Name, label.Color)
	if err != nil {
		return fmt.Errorf("update label: %w", err)
	}
	return nil
}

func (r *LabelRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM label WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	return nil
}

func (r *LabelRepo) AddToEmail(ctx context.Context, userID, emailID, labelID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_label (email_id, label_id)
		 SELECT $1, $2
		 WHERE EXISTS (SELECT 1 FROM email WHERE id = $1 AND user_id = $3)
		   AND EXISTS (SELECT 1 FROM label WHERE id = $2 AND user_id = $3)
		 ON CONFLICT DO NOTHING`,
		emailID, labelID, userID)
	if err != nil {
		return fmt.Errorf("add label to email: %w", err)
	}
	return nil
}

func (r *LabelRepo) RemoveFromEmail(ctx context.Context, userID, emailID, labelID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM email_label
		 WHERE email_id = $1 AND label_id = $2
		   AND EXISTS (SELECT 1 FROM email WHERE id = $1 AND user_id = $3)`,
		emailID, labelID, userID)
	if err != nil {
		return fmt.Errorf("remove label from email: %w", err)
	}
	return nil
}

func (r *LabelRepo) ListByEmail(ctx context.Context, userID, emailID string) ([]domain.Label, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT l.id, l.user_id, l.name, l.color, l.created_at, l.updated_at, 0
		 FROM label l
		 JOIN email_label el ON el.label_id = l.id
		 WHERE el.email_id = $1 AND l.user_id = $2
		 ORDER BY l.name`, emailID, userID)
	if err != nil {
		return nil, fmt.Errorf("list labels by email: %w", err)
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var label domain.Label
		if err := rows.Scan(&label.ID, &label.UserID, &label.Name, &label.Color, &label.CreatedAt, &label.UpdatedAt, &label.EmailCount); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		labels = append(labels, label)
	}
	return labels, nil
}

func (r *LabelRepo) BatchAddToEmails(ctx context.Context, userID string, emailIDs []string, labelID string) error {
	// Verify label belongs to user
	for _, eid := range emailIDs {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO email_label (email_id, label_id)
			 SELECT $1, $2
			 WHERE EXISTS (SELECT 1 FROM email WHERE id = $1 AND user_id = $3)
			   AND EXISTS (SELECT 1 FROM label WHERE id = $2 AND user_id = $3)
			 ON CONFLICT DO NOTHING`,
			eid, labelID, userID)
		if err != nil {
			return fmt.Errorf("batch add label: %w", err)
		}
	}
	return nil
}

func (r *LabelRepo) BatchRemoveFromEmails(ctx context.Context, userID string, emailIDs []string, labelID string) error {
	for _, eid := range emailIDs {
		_, err := r.pool.Exec(ctx,
			`DELETE FROM email_label
			 WHERE email_id = $1 AND label_id = $2
			   AND EXISTS (SELECT 1 FROM email WHERE id = $1 AND user_id = $3)`,
			eid, labelID, userID)
		if err != nil {
			return fmt.Errorf("batch remove label: %w", err)
		}
	}
	return nil
}
