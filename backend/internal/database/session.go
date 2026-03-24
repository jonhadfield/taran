package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT s.id, s."userId", s.token, s."expiresAt", s."updatedAt", u.email
		 FROM session s
		 JOIN "user" u ON u.id = s."userId"
		 WHERE s.token = $1`, token)

	var s domain.Session
	err := row.Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.UpdatedAt, &s.UserEmail)
	if err != nil {
		return nil, fmt.Errorf("get session by token: %w", err)
	}
	return &s, nil
}

// RotateToken replaces the session token with a new value.
func (r *SessionRepo) RotateToken(ctx context.Context, oldToken, newToken string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE session SET token = $1, "updatedAt" = NOW() WHERE token = $2`,
		newToken, oldToken)
	if err != nil {
		return fmt.Errorf("rotate token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (r *SessionRepo) GetUserEmail(ctx context.Context, userID string) (string, error) {
	var email string
	err := r.pool.QueryRow(ctx,
		`SELECT email FROM "user" WHERE id = $1`, userID).Scan(&email)
	if err != nil {
		return "", fmt.Errorf("get user email: %w", err)
	}
	return email, nil
}
