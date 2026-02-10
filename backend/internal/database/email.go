package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailRepo struct {
	pool *pgxpool.Pool
}

func NewEmailRepo(pool *pgxpool.Pool) *EmailRepo {
	return &EmailRepo{pool: pool}
}

func (r *EmailRepo) Create(ctx context.Context, email *domain.Email) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email (id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		email.ID, email.UserID, email.AccountID, email.MessageID,
		email.FromAddress, email.FromName, email.ToAddress, email.Subject,
		email.TextBody, email.HTMLBody, email.ReceivedAt, email.DateHeader,
		email.Status, email.IsRead, email.IsStarred, email.IsArchived,
		email.CreatedAt, email.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create email: %w", err)
	}
	return nil
}

func (r *EmailRepo) GetByID(ctx context.Context, userID, id string) (*domain.Email, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE id = $1 AND user_id = $2`, id, userID)

	return scanEmail(row)
}

func (r *EmailRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Email, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE id = $1`, id)

	return scanEmail(row)
}

func (r *EmailRepo) GetByMessageID(ctx context.Context, messageID string) (*domain.Email, error) {
	if messageID == "" {
		return nil, fmt.Errorf("empty message ID")
	}
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE message_id = $1`, messageID)

	return scanEmail(row)
}

func (r *EmailRepo) List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Email, int, error) {
	where := []string{"user_id = $1"}
	args := []any{userID}
	argIdx := 2

	if opts.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*opts.Status))
		argIdx++
	}
	if opts.IsRead != nil {
		where = append(where, fmt.Sprintf("is_read = $%d", argIdx))
		args = append(args, *opts.IsRead)
		argIdx++
	}
	if opts.IsStarred != nil {
		where = append(where, fmt.Sprintf("is_starred = $%d", argIdx))
		args = append(args, *opts.IsStarred)
		argIdx++
	}
	if opts.IsArchived != nil {
		where = append(where, fmt.Sprintf("is_archived = $%d", argIdx))
		args = append(args, *opts.IsArchived)
		argIdx++
	}
	if opts.Since != nil {
		where = append(where, fmt.Sprintf("received_at >= $%d", argIdx))
		args = append(args, *opts.Since)
		argIdx++
	}
	if opts.Before != nil {
		where = append(where, fmt.Sprintf("received_at < $%d", argIdx))
		args = append(args, *opts.Before)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM email WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count emails: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset

	query := fmt.Sprintf(
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE %s ORDER BY received_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list emails: %w", err)
	}
	defer rows.Close()

	var emails []domain.Email
	for rows.Next() {
		e, err := scanEmailRows(rows)
		if err != nil {
			return nil, 0, err
		}
		emails = append(emails, *e)
	}
	return emails, total, nil
}

func (r *EmailRepo) UpdateState(ctx context.Context, userID, id string, state domain.EmailState) error {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if state.IsRead != nil {
		sets = append(sets, fmt.Sprintf("is_read = $%d", argIdx))
		args = append(args, *state.IsRead)
		argIdx++
	}
	if state.IsStarred != nil {
		sets = append(sets, fmt.Sprintf("is_starred = $%d", argIdx))
		args = append(args, *state.IsStarred)
		argIdx++
	}
	if state.IsArchived != nil {
		sets = append(sets, fmt.Sprintf("is_archived = $%d", argIdx))
		args = append(args, *state.IsArchived)
		argIdx++
	}
	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf("UPDATE email SET %s WHERE id = $%d AND user_id = $%d",
		strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, id, userID)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update email state: %w", err)
	}
	return nil
}

func (r *EmailRepo) ListPending(ctx context.Context, limit int) ([]domain.Email, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE status = 'pending' ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	var emails []domain.Email
	for rows.Next() {
		e, err := scanEmailRows(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *e)
	}
	return emails, nil
}

func (r *EmailRepo) SetStatus(ctx context.Context, id string, status domain.EmailStatus) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE email SET status = $1, updated_at = NOW() WHERE id = $2",
		string(status), id)
	if err != nil {
		return fmt.Errorf("set email status: %w", err)
	}
	return nil
}

func (r *EmailRepo) ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT user_id FROM email WHERE received_at >= $1 AND received_at < $2`,
		from, to)
	if err != nil {
		return nil, fmt.Errorf("list active user IDs: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanEmail(row scannable) (*domain.Email, error) {
	var e domain.Email
	err := row.Scan(
		&e.ID, &e.UserID, &e.AccountID, &e.MessageID,
		&e.FromAddress, &e.FromName, &e.ToAddress, &e.Subject,
		&e.TextBody, &e.HTMLBody, &e.ReceivedAt, &e.DateHeader,
		&e.Status, &e.IsRead, &e.IsStarred, &e.IsArchived,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan email: %w", err)
	}
	return &e, nil
}

func scanEmailRows(rows interface{ Scan(dest ...any) error }) (*domain.Email, error) {
	return scanEmail(rows)
}
