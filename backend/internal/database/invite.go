package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InviteRepo struct {
	pool *pgxpool.Pool
}

func NewInviteRepo(pool *pgxpool.Pool) *InviteRepo {
	return &InviteRepo{pool: pool}
}

func (r *InviteRepo) GetByEmail(ctx context.Context, email string) (*domain.Invite, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email, invited_by, created_at, accepted_at
		 FROM invite WHERE LOWER(email) = LOWER($1)`, email)

	var inv domain.Invite
	err := row.Scan(&inv.ID, &inv.Email, &inv.InvitedBy, &inv.CreatedAt, &inv.AcceptedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get invite by email: %w", err)
	}
	return &inv, nil
}

func (r *InviteRepo) Create(ctx context.Context, invite *domain.Invite) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO invite (id, email, invited_by, created_at)
		 VALUES ($1, $2, $3, $4)`,
		invite.ID, invite.Email, invite.InvitedBy, invite.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (r *InviteRepo) List(ctx context.Context) ([]domain.Invite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, invited_by, created_at, accepted_at
		 FROM invite ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	defer rows.Close()

	var invites []domain.Invite
	for rows.Next() {
		var inv domain.Invite
		if err := rows.Scan(&inv.ID, &inv.Email, &inv.InvitedBy, &inv.CreatedAt, &inv.AcceptedAt); err != nil {
			return nil, fmt.Errorf("scan invite: %w", err)
		}
		invites = append(invites, inv)
	}
	return invites, nil
}

func (r *InviteRepo) MarkAccepted(ctx context.Context, email string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE invite SET accepted_at = $1 WHERE LOWER(email) = LOWER($2) AND accepted_at IS NULL`,
		now, email,
	)
	if err != nil {
		return fmt.Errorf("mark invite accepted: %w", err)
	}
	return nil
}
