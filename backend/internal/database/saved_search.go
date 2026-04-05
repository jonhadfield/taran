package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SavedSearchRepo struct {
	pool *pgxpool.Pool
}

func NewSavedSearchRepo(pool *pgxpool.Pool) *SavedSearchRepo {
	return &SavedSearchRepo{pool: pool}
}

func (r *SavedSearchRepo) Create(ctx context.Context, s *domain.SavedSearch) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO saved_search (id, user_id, name, filters, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		s.ID, s.UserID, s.Name, s.Filters)
	if err != nil {
		return fmt.Errorf("create saved search: %w", err)
	}
	return nil
}

func (r *SavedSearchRepo) ListByUser(ctx context.Context, userID string) ([]domain.SavedSearch, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, filters, created_at, updated_at
		 FROM saved_search WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list saved searches: %w", err)
	}
	defer rows.Close()

	var searches []domain.SavedSearch
	for rows.Next() {
		var s domain.SavedSearch
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Filters, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan saved search: %w", err)
		}
		searches = append(searches, s)
	}
	return searches, nil
}

func (r *SavedSearchRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM saved_search WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete saved search: %w", err)
	}
	return nil
}

func (r *SavedSearchRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM saved_search WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count saved searches: %w", err)
	}
	return int(count), nil
}
