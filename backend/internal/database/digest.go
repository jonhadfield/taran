package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDigestDuplicate is returned when a digest already exists for the same
// user/period_start/period_end combination (unique constraint violation).
var ErrDigestDuplicate = errors.New("digest already exists for this period")

type DigestRepo struct {
	pool *pgxpool.Pool
}

func NewDigestRepo(pool *pgxpool.Pool) *DigestRepo {
	return &DigestRepo{pool: pool}
}

func (r *DigestRepo) Create(ctx context.Context, digest *domain.Digest) error {
	highlights, _ := json.Marshal(digest.Highlights)
	topTopics, _ := json.Marshal(digest.TopTopics)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO digest (id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		digest.ID, digest.UserID, digest.Title, digest.Summary,
		highlights, topTopics,
		digest.PeriodStart, digest.PeriodEnd, digest.PeriodType,
		digest.EmailCount, digest.TokensUsed, digest.Provider, digest.Model,
		digest.GeneratedAt, digest.SentAt, digest.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDigestDuplicate
		}
		return fmt.Errorf("insert digest: %w", err)
	}

	for _, item := range digest.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO digest_item (id, digest_id, email_id, extraction_id, sort_order, created_at)
			 VALUES ($1,$2,$3,$4,$5,NOW())`,
			item.ID, digest.ID, item.EmailID, item.ExtractionID, item.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert digest item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *DigestRepo) GetByID(ctx context.Context, userID, id string) (*domain.Digest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at, share_token, share_token_created_at
		 FROM digest WHERE id = $1 AND user_id = $2`, id, userID)

	d, err := scanDigest(row)
	if err != nil {
		return nil, err
	}

	items, err := r.queryDigestItems(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Items = items

	return d, nil
}

func (r *DigestRepo) GetByIDInternal(ctx context.Context, id string) (*domain.Digest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at, share_token, share_token_created_at
		 FROM digest WHERE id = $1`, id)

	d, err := scanDigest(row)
	if err != nil {
		return nil, err
	}

	items, err := r.queryDigestItems(ctx, id)
	if err != nil {
		return nil, err
	}
	d.Items = items

	return d, nil
}

func (r *DigestRepo) List(ctx context.Context, userID string, opts domain.ListOptions) ([]domain.Digest, int, error) {
	where := []string{"user_id = $1"}
	args := []any{userID}
	argIdx := 2

	if opts.Since != nil {
		where = append(where, fmt.Sprintf("period_start >= $%d", argIdx))
		args = append(args, *opts.Since)
		argIdx++
	}
	if opts.Before != nil {
		where = append(where, fmt.Sprintf("period_end < $%d", argIdx))
		args = append(args, *opts.Before)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM digest WHERE "+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count digests: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at, share_token, share_token_created_at
		 FROM digest WHERE %s ORDER BY period_start DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, limit, opts.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list digests: %w", err)
	}
	defer rows.Close()

	var digests []domain.Digest
	for rows.Next() {
		d, err := scanDigest(rows)
		if err != nil {
			return nil, 0, err
		}
		digests = append(digests, *d)
	}
	return digests, total, nil
}

func (r *DigestRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM digest WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("delete digest: %w", err)
	}
	return nil
}

func (r *DigestRepo) SetSentAt(ctx context.Context, id string, sentAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE digest SET sent_at = $1 WHERE id = $2`, sentAt, id)
	if err != nil {
		return fmt.Errorf("set sent_at: %w", err)
	}
	return nil
}

func (r *DigestRepo) SetShareToken(ctx context.Context, id, userID, token string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE digest SET share_token = $1, share_token_created_at = NOW() WHERE id = $2 AND user_id = $3`,
		token, id, userID)
	if err != nil {
		return fmt.Errorf("set share token: %w", err)
	}
	return nil
}

func (r *DigestRepo) ClearShareToken(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE digest SET share_token = NULL, share_token_created_at = NULL WHERE id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		return fmt.Errorf("clear share token: %w", err)
	}
	return nil
}

func (r *DigestRepo) GetByShareToken(ctx context.Context, token string) (*domain.Digest, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at, share_token, share_token_created_at
		 FROM digest WHERE share_token = $1`, token)

	d, err := scanDigest(row)
	if err != nil {
		return nil, err
	}

	items, err := r.queryDigestItems(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	d.Items = items

	return d, nil
}

func (r *DigestRepo) ExistsForPeriod(ctx context.Context, userID string, periodStart, periodEnd time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM digest WHERE user_id = $1 AND period_start = $2 AND period_end = $3)`,
		userID, periodStart, periodEnd,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check digest exists for period: %w", err)
	}
	return exists, nil
}

func (r *DigestRepo) ListUnsent(ctx context.Context, olderThan time.Time, limit int) ([]domain.Digest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, summary, highlights, top_topics,
		    period_start, period_end, period_type, email_count, tokens_used, provider, model,
		    generated_at, sent_at, created_at, share_token, share_token_created_at
		 FROM digest
		 WHERE sent_at IS NULL AND generated_at < $1
		 ORDER BY generated_at ASC
		 LIMIT $2`, olderThan, limit)
	if err != nil {
		return nil, fmt.Errorf("list unsent digests: %w", err)
	}
	defer rows.Close()

	var digests []domain.Digest
	for rows.Next() {
		d, err := scanDigest(rows)
		if err != nil {
			return nil, err
		}
		digests = append(digests, *d)
	}
	return digests, nil
}

func (r *DigestRepo) queryDigestItems(ctx context.Context, digestID string) ([]domain.DigestItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT di.id, di.digest_id, di.email_id, di.extraction_id, di.sort_order,
		        e.subject, e.from_name, e.from_address,
		        COALESCE(ex.summary, '')
		 FROM digest_item di
		 JOIN email e ON e.id = di.email_id
		 LEFT JOIN extraction ex ON ex.id = di.extraction_id
		 WHERE di.digest_id = $1
		 ORDER BY di.sort_order`, digestID)
	if err != nil {
		return nil, fmt.Errorf("query digest items: %w", err)
	}
	defer rows.Close()

	var items []domain.DigestItem
	for rows.Next() {
		var item domain.DigestItem
		if err := rows.Scan(
			&item.ID, &item.DigestID, &item.EmailID, &item.ExtractionID, &item.SortOrder,
			&item.Subject, &item.FromName, &item.FromAddress, &item.Summary,
		); err != nil {
			return nil, fmt.Errorf("scan digest item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func scanDigest(row scannable) (*domain.Digest, error) {
	var d domain.Digest
	var highlights, topTopics []byte

	err := row.Scan(
		&d.ID, &d.UserID, &d.Title, &d.Summary, &highlights, &topTopics,
		&d.PeriodStart, &d.PeriodEnd, &d.PeriodType, &d.EmailCount,
		&d.TokensUsed, &d.Provider, &d.Model, &d.GeneratedAt, &d.SentAt, &d.CreatedAt,
		&d.ShareToken, &d.ShareTokenCreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan digest: %w", err)
	}

	json.Unmarshal(highlights, &d.Highlights)
	json.Unmarshal(topTopics, &d.TopTopics)

	return &d, nil
}

// DeleteOrphaned removes digests that have no remaining digest_items.
func (r *DigestRepo) DeleteOrphaned(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM digest WHERE id IN (
			SELECT d.id FROM digest d
			LEFT JOIN digest_item di ON di.digest_id = d.id
			GROUP BY d.id
			HAVING COUNT(di.id) = 0
		)`)
	if err != nil {
		return 0, fmt.Errorf("delete orphaned digests: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
