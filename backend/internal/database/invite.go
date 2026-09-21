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
		 FROM invite ORDER BY created_at DESC LIMIT 200`)
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

// CountByInviter returns how many invites were created by the given inviter,
// e.g. auth.OpenRegistrationInviter for open-registration sign-ups.
func (r *InviteRepo) CountByInviter(ctx context.Context, invitedBy string) (int, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM invite WHERE invited_by = $1`, invitedBy).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count invites by inviter: %w", err)
	}
	return int(count), nil
}

func (r *InviteRepo) MarkAccepted(ctx context.Context, email string) (bool, error) {
	now := time.Now().UTC()
	// The accepted_at IS NULL guard makes this a one-time transition, so
	// concurrent callers can't both see themselves as the first acceptance.
	tag, err := r.pool.Exec(ctx,
		`UPDATE invite SET accepted_at = $1 WHERE LOWER(email) = LOWER($2) AND accepted_at IS NULL`,
		now, email,
	)
	if err != nil {
		return false, fmt.Errorf("mark invite accepted: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
