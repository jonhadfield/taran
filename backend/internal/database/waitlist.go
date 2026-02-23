package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WaitlistRepo struct {
	pool *pgxpool.Pool
}

func NewWaitlistRepo(pool *pgxpool.Pool) *WaitlistRepo {
	return &WaitlistRepo{pool: pool}
}

func (r *WaitlistRepo) Create(ctx context.Context, req *domain.WaitlistRequest) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO waitlist_request (id, email, created_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO NOTHING`,
		req.ID, req.Email, req.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create waitlist request: %w", err)
	}
	return nil
}

func (r *WaitlistRepo) GetByEmail(ctx context.Context, email string) (*domain.WaitlistRequest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, created_at
		 FROM waitlist_request WHERE LOWER(email) = LOWER($1)`, email)

	var req domain.WaitlistRequest
	err := row.Scan(&req.ID, &req.Email, &req.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get waitlist request by email: %w", err)
	}
	return &req, nil
}

func (r *WaitlistRepo) List(ctx context.Context) ([]domain.WaitlistRequest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, created_at
		 FROM waitlist_request ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("list waitlist requests: %w", err)
	}
	defer rows.Close()

	var requests []domain.WaitlistRequest
	for rows.Next() {
		var req domain.WaitlistRequest
		if err := rows.Scan(&req.ID, &req.Email, &req.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan waitlist request: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (r *WaitlistRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM waitlist_request WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete waitlist request: %w", err)
	}
	return nil
}
