package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SenderPreferenceRepo struct {
	pool *pgxpool.Pool
}

func NewSenderPreferenceRepo(pool *pgxpool.Pool) *SenderPreferenceRepo {
	return &SenderPreferenceRepo{pool: pool}
}

func (r *SenderPreferenceRepo) Upsert(ctx context.Context, pref *domain.SenderPreference) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sender_preference (id, user_id, from_address, status, category, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, from_address) DO UPDATE SET status = $4, category = $5, updated_at = $7`,
		pref.ID, pref.UserID, pref.FromAddress, pref.Status, pref.Category, pref.CreatedAt, pref.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert sender preference: %w", err)
	}
	return nil
}

func (r *SenderPreferenceRepo) GetByAddress(ctx context.Context, userID, fromAddress string) (*domain.SenderPreference, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, from_address, status, category, unsubscribed_at, created_at, updated_at
		 FROM sender_preference WHERE user_id = $1 AND from_address = $2`,
		userID, fromAddress)

	var p domain.SenderPreference
	err := row.Scan(&p.ID, &p.UserID, &p.FromAddress, &p.Status, &p.Category, &p.UnsubscribedAt, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sender preference: %w", err)
	}
	return &p, nil
}

func (r *SenderPreferenceRepo) ListByUser(ctx context.Context, userID string) ([]domain.SenderPreference, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, from_address, status, category, unsubscribed_at, created_at, updated_at
		 FROM sender_preference WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list sender preferences: %w", err)
	}
	defer rows.Close()

	var prefs []domain.SenderPreference
	for rows.Next() {
		var p domain.SenderPreference
		if err := rows.Scan(&p.ID, &p.UserID, &p.FromAddress, &p.Status, &p.Category, &p.UnsubscribedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sender preference: %w", err)
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (r *SenderPreferenceRepo) ListBlockedAddresses(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT from_address FROM sender_preference WHERE user_id = $1 AND status = 'blocked'`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list blocked addresses: %w", err)
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, fmt.Errorf("scan blocked address: %w", err)
		}
		addresses = append(addresses, addr)
	}
	return addresses, nil
}

func (r *SenderPreferenceRepo) MarkUnsubscribed(ctx context.Context, userID, fromAddress string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sender_preference (id, user_id, from_address, status, category, unsubscribed_at, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, 'normal', '', NOW(), NOW(), NOW())
		 ON CONFLICT (user_id, from_address) DO UPDATE SET unsubscribed_at = NOW(), updated_at = NOW()`,
		userID, fromAddress)
	if err != nil {
		return fmt.Errorf("mark unsubscribed: %w", err)
	}
	return nil
}
