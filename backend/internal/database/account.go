package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepo struct {
	pool *pgxpool.Pool
}

func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{pool: pool}
}

func (r *AccountRepo) Create(ctx context.Context, account *domain.EmailAccount) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_account (id, user_id, email_address, display_name, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		account.ID, account.UserID, account.EmailAddress, account.DisplayName, account.IsActive, account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

func (r *AccountRepo) GetByEmailAddress(ctx context.Context, emailAddress string) (*domain.EmailAccount, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_address, display_name, is_active, created_at, updated_at
		 FROM email_account WHERE email_address = $1`, emailAddress)

	var a domain.EmailAccount
	err := row.Scan(&a.ID, &a.UserID, &a.EmailAddress, &a.DisplayName, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get account by email: %w", err)
	}
	return &a, nil
}

func (r *AccountRepo) GetByID(ctx context.Context, userID, id string) (*domain.EmailAccount, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_address, display_name, is_active, created_at, updated_at
		 FROM email_account WHERE id = $1 AND user_id = $2`, id, userID)

	var a domain.EmailAccount
	err := row.Scan(&a.ID, &a.UserID, &a.EmailAddress, &a.DisplayName, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}
	return &a, nil
}

func (r *AccountRepo) ListByUser(ctx context.Context, userID string) ([]domain.EmailAccount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email_address, display_name, is_active, created_at, updated_at
		 FROM email_account WHERE user_id = $1 AND is_active = true ORDER BY created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []domain.EmailAccount
	for rows.Next() {
		var a domain.EmailAccount
		if err := rows.Scan(&a.ID, &a.UserID, &a.EmailAddress, &a.DisplayName, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *AccountRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE email_account SET is_active = false, updated_at = NOW() WHERE id = $1 AND user_id = $2 AND is_active = true`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}
