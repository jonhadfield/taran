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
		`SELECT s.id, s."userId", s.token, s."expiresAt", u.email
		 FROM session s
		 JOIN "user" u ON u.id = s."userId"
		 WHERE s.token = $1`, token)

	var s domain.Session
	err := row.Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.UserEmail)
	if err != nil {
		return nil, fmt.Errorf("get session by token: %w", err)
	}
	return &s, nil
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
