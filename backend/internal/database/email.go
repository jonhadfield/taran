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
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		email.ID, email.UserID, email.AccountID, email.MessageID,
		email.FromAddress, email.FromName, email.ToAddress, email.Subject,
		email.TextBody, email.HTMLBody, email.ReceivedAt, email.DateHeader,
		email.Status, email.StatusReason, email.IsRead, email.IsStarred, email.IsArchived,
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
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE id = $1 AND user_id = $2`, id, userID)

	return scanEmail(row)
}

func (r *EmailRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Email, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE id = $1`, id)

	return scanEmail(row)
}

func (r *EmailRepo) GetByIDsInternal(ctx context.Context, ids []string) ([]domain.Email, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, created_at, updated_at
		 FROM email WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("get emails by ids: %w", err)
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

func (r *EmailRepo) GetByMessageID(ctx context.Context, messageID string) (*domain.Email, error) {
	if messageID == "" {
		return nil, fmt.Errorf("empty message ID")
	}
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
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
	if opts.Search != nil && *opts.Search != "" {
		searchTerm := "%" + *opts.Search + "%"
		where = append(where, fmt.Sprintf(
			"(subject ILIKE $%d OR from_name ILIKE $%d OR from_address ILIKE $%d OR EXISTS (SELECT 1 FROM extraction ex WHERE ex.email_id = email.id AND ex.summary ILIKE $%d))",
			argIdx, argIdx+1, argIdx+2, argIdx+3))
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		argIdx += 4
	}
	if opts.Topic != nil && *opts.Topic != "" {
		topicJSON := fmt.Sprintf(`[%q]`, *opts.Topic)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM extraction ex WHERE ex.email_id = email.id AND ex.topics @> $%d::jsonb)",
			argIdx))
		args = append(args, topicJSON)
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
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
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

func (r *EmailRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM email WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("delete email: %w", err)
	}
	return nil
}

func (r *EmailRepo) BatchUpdateState(ctx context.Context, userID string, ids []string, state domain.EmailState) error {
	if len(ids) == 0 {
		return nil
	}

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

	sets = append(sets, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE email SET %s WHERE id = ANY($%d) AND user_id = $%d",
		strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, ids, userID)

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch update email state: %w", err)
	}
	return nil
}

func (r *EmailRepo) BatchDelete(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		"DELETE FROM email WHERE id = ANY($1) AND user_id = $2", ids, userID)
	if err != nil {
		return fmt.Errorf("batch delete emails: %w", err)
	}
	return nil
}

func (r *EmailRepo) CountByWeek(ctx context.Context, userID string, weeks int) ([]domain.WeekCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT date_trunc('week', received_at) AS week, COUNT(*)
		 FROM email
		 WHERE user_id = $1 AND received_at > NOW() - ($2 * INTERVAL '1 week')
		 GROUP BY 1 ORDER BY 1`,
		userID, weeks)
	if err != nil {
		return nil, fmt.Errorf("count by week: %w", err)
	}
	defer rows.Close()

	var result []domain.WeekCount
	for rows.Next() {
		var wc domain.WeekCount
		if err := rows.Scan(&wc.Week, &wc.Count); err != nil {
			return nil, fmt.Errorf("scan week count: %w", err)
		}
		result = append(result, wc)
	}
	return result, nil
}

func (r *EmailRepo) ListPending(ctx context.Context, limit int) ([]domain.Email, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, email_account_id, message_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
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

func (r *EmailRepo) SetStatus(ctx context.Context, id string, status domain.EmailStatus, reason string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE email SET status = $1, status_reason = $2, updated_at = NOW() WHERE id = $3",
		string(status), reason, id)
	if err != nil {
		return fmt.Errorf("set email status: %w", err)
	}
	return nil
}

func (r *EmailRepo) SetStatusScoped(ctx context.Context, userID, id string, status domain.EmailStatus, reason string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE email SET status = $1, status_reason = $2, updated_at = NOW() WHERE id = $3 AND user_id = $4",
		string(status), reason, id, userID)
	if err != nil {
		return fmt.Errorf("set email status (scoped): %w", err)
	}
	return nil
}

func (r *EmailRepo) ListActiveUserIDs(ctx context.Context, from, to time.Time) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT user_id FROM email WHERE received_at >= $1 AND received_at < $2 AND status = 'processed'`,
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

func (r *EmailRepo) CountByPeriod(ctx context.Context, userID string, from, to time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM email WHERE user_id = $1 AND received_at >= $2 AND received_at < $3 AND status = 'processed'",
		userID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count by period: %w", err)
	}
	return count, nil
}

func (r *EmailRepo) TopSenders(ctx context.Context, userID string, from, to time.Time, limit int) ([]domain.SenderCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT from_address, from_name, COUNT(*) as cnt
		 FROM email WHERE user_id = $1 AND received_at >= $2 AND received_at < $3 AND status = 'processed'
		 GROUP BY from_address, from_name ORDER BY cnt DESC LIMIT $4`,
		userID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("top senders: %w", err)
	}
	defer rows.Close()

	var senders []domain.SenderCount
	for rows.Next() {
		var s domain.SenderCount
		if err := rows.Scan(&s.FromAddress, &s.FromName, &s.Count); err != nil {
			return nil, fmt.Errorf("scan sender count: %w", err)
		}
		senders = append(senders, s)
	}
	return senders, nil
}

func (r *EmailRepo) ListSenders(ctx context.Context, userID string) ([]domain.SenderInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT from_address, COALESCE(MAX(from_name), '') as from_name, COUNT(*) as cnt
		 FROM email WHERE user_id = $1 GROUP BY from_address ORDER BY cnt DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list senders: %w", err)
	}
	defer rows.Close()

	var senders []domain.SenderInfo
	for rows.Next() {
		var s domain.SenderInfo
		if err := rows.Scan(&s.FromAddress, &s.FromName, &s.EmailCount); err != nil {
			return nil, fmt.Errorf("scan sender info: %w", err)
		}
		senders = append(senders, s)
	}
	return senders, nil
}

func (r *EmailRepo) CountByStatus(ctx context.Context, userID string) (map[domain.EmailStatus]int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM email WHERE user_id = $1 GROUP BY status`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("count by status: %w", err)
	}
	defer rows.Close()

	counts := make(map[domain.EmailStatus]int)
	for rows.Next() {
		var status domain.EmailStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan status count: %w", err)
		}
		counts[status] = count
	}
	return counts, nil
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
		&e.Status, &e.StatusReason, &e.IsRead, &e.IsStarred, &e.IsArchived,
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
