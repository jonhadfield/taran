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
		`INSERT INTO email (id, user_id, email_account_id, message_id, in_reply_to, thread_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, unsubscribe_url, unsubscribe_mailto, retry_count, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		email.ID, email.UserID, email.AccountID, email.MessageID, email.InReplyTo, email.ThreadID,
		email.FromAddress, email.FromName, email.ToAddress, email.Subject,
		email.TextBody, email.HTMLBody, email.ReceivedAt, email.DateHeader,
		email.Status, email.StatusReason, email.IsRead, email.IsStarred, email.IsArchived,
		email.UnsubscribeURL, email.UnsubscribeMailto, email.RetryCount,
		email.CreatedAt, email.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create email: %w", err)
	}
	return nil
}

func (r *EmailRepo) GetByID(ctx context.Context, userID, id string) (*domain.Email, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT ` + emailColumns + `
		 FROM email WHERE id = $1 AND user_id = $2`, id, userID)

	return scanEmail(row)
}

func (r *EmailRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Email, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT ` + emailColumns + `
		 FROM email WHERE id = $1`, id)

	return scanEmail(row)
}

func (r *EmailRepo) GetByIDsInternal(ctx context.Context, ids []string) ([]domain.Email, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT ` + emailColumns + `
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
		`SELECT ` + emailColumns + `
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
		// Full-text search on tsvector column, ILIKE fallback on from_address/from_name,
		// and search extraction summary/key_points/action_items for AI-extracted content
		where = append(where, fmt.Sprintf(
			`(search_tsv @@ plainto_tsquery('english', $%d)
			 OR from_address ILIKE $%d
			 OR from_name ILIKE $%d
			 OR EXISTS (SELECT 1 FROM extraction ex WHERE ex.email_id = email.id
			   AND (ex.summary ILIKE $%d
			     OR EXISTS (SELECT 1 FROM jsonb_array_elements_text(ex.key_points) kp WHERE kp ILIKE $%d)
			     OR EXISTS (SELECT 1 FROM jsonb_array_elements_text(ex.action_items) ai WHERE ai ILIKE $%d))))`,
			argIdx, argIdx+1, argIdx+1, argIdx+1, argIdx+1, argIdx+1))
		args = append(args, *opts.Search, "%"+*opts.Search+"%")
		argIdx += 2
	}
	if opts.Topic != nil && *opts.Topic != "" {
		topicJSON := fmt.Sprintf(`[%q]`, *opts.Topic)
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM extraction ex WHERE ex.email_id = email.id AND ex.topics @> $%d::jsonb)",
			argIdx))
		args = append(args, topicJSON)
		argIdx++
	}
	if opts.Category != nil && *opts.Category != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM extraction ex WHERE ex.email_id = email.id AND ex.source_category = $%d)",
			argIdx))
		args = append(args, *opts.Category)
		argIdx++
	}
	if opts.FromAddress != nil && *opts.FromAddress != "" {
		where = append(where, fmt.Sprintf("from_address = $%d", argIdx))
		args = append(args, *opts.FromAddress)
		argIdx++
	}
	if opts.HasAttachment != nil && *opts.HasAttachment {
		where = append(where, "EXISTS (SELECT 1 FROM email_attachment ea WHERE ea.email_id = email.id)")
	}
	if opts.LabelID != nil && *opts.LabelID != "" {
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM email_label el WHERE el.email_id = email.id AND el.label_id = $%d)",
			argIdx))
		args = append(args, *opts.LabelID)
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

	orderBy := "received_at DESC"
	switch opts.Sort {
	case "oldest":
		orderBy = "received_at ASC"
	case "relevance":
		if opts.Search != nil && *opts.Search != "" {
			orderBy = fmt.Sprintf("ts_rank(search_tsv, plainto_tsquery('english', $%d)) DESC, received_at DESC", argIdx)
			args = append(args, *opts.Search)
			argIdx++
		}
	default: // "newest" or empty
		if opts.Search != nil && *opts.Search != "" {
			// Default to relevance when searching
			orderBy = fmt.Sprintf("ts_rank(search_tsv, plainto_tsquery('english', $%d)) DESC, received_at DESC", argIdx)
			args = append(args, *opts.Search)
			argIdx++
		}
	}

	query := fmt.Sprintf(
		`SELECT ` + emailColumns + `
		 FROM email WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIdx, argIdx+1)
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
		var count int64
		if err := rows.Scan(&wc.Week, &count); err != nil {
			return nil, fmt.Errorf("scan week count: %w", err)
		}
		wc.Count = int(count)
		result = append(result, wc)
	}
	return result, nil
}

func (r *EmailRepo) ListPending(ctx context.Context, limit int) ([]domain.Email, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ` + emailColumns + `
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

func (r *EmailRepo) ListRetryable(ctx context.Context, maxRetries, limit int) ([]domain.Email, error) {
	// Find failed emails where:
	// 1. retry_count < maxRetries
	// 2. enough time has passed since last update (exponential backoff: 5min, 20min, 80min, ...)
	// 3. failure reason is retryable (not "email has no content", "sender is blocked")
	rows, err := r.pool.Query(ctx,
		`SELECT ` + emailColumns + `
		 FROM email
		 WHERE status = 'failed'
		   AND retry_count < $1
		   AND status_reason NOT IN ('email has no content', 'sender is blocked')
		   AND updated_at < NOW() - (INTERVAL '5 minutes' * POWER(4, retry_count))
		 ORDER BY updated_at ASC
		 LIMIT $2`, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("list retryable: %w", err)
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

func (r *EmailRepo) IncrementRetryCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE email SET retry_count = retry_count + 1, updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("increment retry count: %w", err)
	}
	return nil
}

func (r *EmailRepo) ResetRetryCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE email SET retry_count = 0, updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("reset retry count: %w", err)
	}
	return nil
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
		var cnt int64
		if err := rows.Scan(&s.FromAddress, &s.FromName, &cnt); err != nil {
			return nil, fmt.Errorf("scan sender count: %w", err)
		}
		s.Count = int(cnt)
		senders = append(senders, s)
	}
	return senders, nil
}

func (r *EmailRepo) ListSenders(ctx context.Context, userID string) ([]domain.SenderInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.from_address,
		        COALESCE(MAX(e.from_name), '') as from_name,
		        COUNT(*) as cnt,
		        COALESCE((
		            SELECT ex.source_category
		            FROM extraction ex
		            JOIN email e2 ON e2.id = ex.email_id
		            WHERE e2.user_id = $1 AND e2.from_address = e.from_address
		              AND ex.source_category != ''
		            GROUP BY ex.source_category
		            ORDER BY COUNT(*) DESC
		            LIMIT 1
		        ), '') as auto_category
		 FROM email e WHERE e.user_id = $1 GROUP BY e.from_address ORDER BY cnt DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list senders: %w", err)
	}
	defer rows.Close()

	var senders []domain.SenderInfo
	for rows.Next() {
		var s domain.SenderInfo
		if err := rows.Scan(&s.FromAddress, &s.FromName, &s.EmailCount, &s.AutoCategory); err != nil {
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

func (r *EmailRepo) CountByHourAndDay(ctx context.Context, userID string) ([]domain.HeatmapCell, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT EXTRACT(HOUR FROM received_at)::int AS hour,
		        EXTRACT(DOW FROM received_at)::int AS day,
		        COUNT(*)::int AS cnt
		 FROM email WHERE user_id = $1
		 GROUP BY hour, day ORDER BY day, hour`, userID)
	if err != nil {
		return nil, fmt.Errorf("count by hour and day: %w", err)
	}
	defer rows.Close()

	var cells []domain.HeatmapCell
	for rows.Next() {
		var c domain.HeatmapCell
		var hour, day, cnt int64
		if err := rows.Scan(&hour, &day, &cnt); err != nil {
			return nil, fmt.Errorf("scan heatmap cell: %w", err)
		}
		c.Hour = int(hour)
		c.Day = int(day)
		c.Count = int(cnt)
		cells = append(cells, c)
	}
	return cells, nil
}

func (r *EmailRepo) ListSubscriptions(ctx context.Context, userID string) ([]domain.SubscriptionInfo, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.from_address,
		        COALESCE(MAX(e.from_name), '') AS from_name,
		        COUNT(*) AS cnt,
		        MAX(e.received_at) AS last_seen,
		        COALESCE(MAX(e.unsubscribe_url), '') AS unsub_url,
		        COALESCE(MAX(e.unsubscribe_mailto), '') AS unsub_mailto
		 FROM email e
		 WHERE e.user_id = $1
		   AND (e.unsubscribe_url != '' OR e.unsubscribe_mailto != '')
		 GROUP BY e.from_address
		 ORDER BY last_seen DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []domain.SubscriptionInfo
	for rows.Next() {
		var s domain.SubscriptionInfo
		var cnt int64
		if err := rows.Scan(&s.FromAddress, &s.FromName, &cnt, &s.LastSeen, &s.UnsubscribeURL, &s.UnsubscribeMailto); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		s.EmailCount = int(cnt)
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *EmailRepo) GetSenderDetail(ctx context.Context, userID, fromAddress string) (*domain.SenderDetail, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT e.from_address,
		        COALESCE(MAX(e.from_name), '') as from_name,
		        COUNT(*) as cnt,
		        MIN(e.received_at) as first_seen,
		        MAX(e.received_at) as last_seen,
		        COALESCE((
		            SELECT ex.source_category
		            FROM extraction ex
		            JOIN email e2 ON e2.id = ex.email_id
		            WHERE e2.user_id = $1 AND e2.from_address = $2
		              AND ex.source_category != ''
		            GROUP BY ex.source_category
		            ORDER BY COUNT(*) DESC
		            LIMIT 1
		        ), '') as auto_category
		 FROM email e WHERE e.user_id = $1 AND e.from_address = $2
		 GROUP BY e.from_address`,
		userID, fromAddress)

	var d domain.SenderDetail
	err := row.Scan(&d.FromAddress, &d.FromName, &d.EmailCount, &d.FirstSeen, &d.LastSeen, &d.AutoCategory)
	if err != nil {
		return nil, fmt.Errorf("get sender detail: %w", err)
	}
	return &d, nil
}

func (r *EmailRepo) CountBySenderWeek(ctx context.Context, userID, fromAddress string, weeks int) ([]domain.WeekCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT date_trunc('week', received_at) AS week, COUNT(*)
		 FROM email
		 WHERE user_id = $1 AND from_address = $2 AND received_at > NOW() - ($3 * INTERVAL '1 week')
		 GROUP BY 1 ORDER BY 1`,
		userID, fromAddress, weeks)
	if err != nil {
		return nil, fmt.Errorf("count by sender week: %w", err)
	}
	defer rows.Close()

	var result []domain.WeekCount
	for rows.Next() {
		var wc domain.WeekCount
		var count int64
		if err := rows.Scan(&wc.Week, &count); err != nil {
			return nil, fmt.Errorf("scan week count: %w", err)
		}
		wc.Count = int(count)
		result = append(result, wc)
	}
	return result, nil
}

func (r *EmailRepo) ListFailedAll(ctx context.Context, limit, offset int) ([]domain.FailedEmail, int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM email WHERE status IN ('failed', 'processing')`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count failed emails: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.user_id, e.email_account_id, e.message_id, e.from_address, e.from_name,
		    e.to_address, e.subject, e.text_body, e.html_body, e.received_at, e.date_header,
		    e.status, e.status_reason, e.is_read, e.is_starred, e.is_archived,
		    e.unsubscribe_url, e.unsubscribe_mailto, e.retry_count, e.created_at, e.updated_at,
		    COALESCE(u.email, '') as user_email
		 FROM email e
		 LEFT JOIN "user" u ON u.id = e.user_id
		 WHERE e.status IN ('failed', 'processing')
		 ORDER BY e.updated_at DESC
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list failed emails: %w", err)
	}
	defer rows.Close()

	var emails []domain.FailedEmail
	for rows.Next() {
		var fe domain.FailedEmail
		err := rows.Scan(
			&fe.ID, &fe.UserID, &fe.AccountID, &fe.MessageID,
			&fe.FromAddress, &fe.FromName, &fe.ToAddress, &fe.Subject,
			&fe.TextBody, &fe.HTMLBody, &fe.ReceivedAt, &fe.DateHeader,
			&fe.Status, &fe.StatusReason, &fe.IsRead, &fe.IsStarred, &fe.IsArchived,
			&fe.UnsubscribeURL, &fe.UnsubscribeMailto, &fe.RetryCount,
			&fe.CreatedAt, &fe.UpdatedAt, &fe.UserEmail,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan failed email: %w", err)
		}
		emails = append(emails, fe)
	}
	return emails, total, nil
}

func (r *EmailRepo) BatchResetForRetry(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE email SET status = 'pending', retry_count = 0, status_reason = '', updated_at = NOW()
		 WHERE id = ANY($1)`, ids)
	if err != nil {
		return 0, fmt.Errorf("batch reset for retry: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *EmailRepo) DeleteInternal(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM email WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete email (internal): %w", err)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

const emailColumns = `id, user_id, email_account_id, message_id, in_reply_to, thread_id, from_address, from_name,
		    to_address, subject, text_body, html_body, received_at, date_header, status, status_reason,
		    is_read, is_starred, is_archived, unsubscribe_url, unsubscribe_mailto, retry_count, created_at, updated_at`

func scanEmail(row scannable) (*domain.Email, error) {
	var e domain.Email
	err := row.Scan(
		&e.ID, &e.UserID, &e.AccountID, &e.MessageID, &e.InReplyTo, &e.ThreadID,
		&e.FromAddress, &e.FromName, &e.ToAddress, &e.Subject,
		&e.TextBody, &e.HTMLBody, &e.ReceivedAt, &e.DateHeader,
		&e.Status, &e.StatusReason, &e.IsRead, &e.IsStarred, &e.IsArchived,
		&e.UnsubscribeURL, &e.UnsubscribeMailto, &e.RetryCount,
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

func (r *EmailRepo) GetThreadEmails(ctx context.Context, userID, threadID string) ([]domain.Email, error) {
	if threadID == "" {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+emailColumns+`
		 FROM email WHERE user_id = $1 AND thread_id = $2
		 ORDER BY received_at ASC`, userID, threadID)
	if err != nil {
		return nil, fmt.Errorf("get thread emails: %w", err)
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

func (r *EmailRepo) UpdateThreadID(ctx context.Context, id, threadID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE email SET thread_id = $1, updated_at = NOW() WHERE id = $2`,
		threadID, id)
	if err != nil {
		return fmt.Errorf("update thread id: %w", err)
	}
	return nil
}
