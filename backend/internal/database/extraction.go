package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExtractionRepo struct {
	pool *pgxpool.Pool
}

func NewExtractionRepo(pool *pgxpool.Pool) *ExtractionRepo {
	return &ExtractionRepo{pool: pool}
}

func (r *ExtractionRepo) Create(ctx context.Context, extraction *domain.Extraction) error {
	keyPoints, _ := json.Marshal(extraction.KeyPoints)
	topics, _ := json.Marshal(extraction.Topics)
	links, _ := json.Marshal(extraction.Links)

	_, err := r.pool.Exec(ctx,
		`INSERT INTO extraction (id, email_id, summary, key_points, topics, links,
		    sentiment, source_category, provider, model, tokens_used, processed_at, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		extraction.ID, extraction.EmailID, extraction.Summary,
		keyPoints, topics, links,
		extraction.Sentiment, extraction.SourceCategory,
		extraction.Provider, extraction.Model, extraction.TokensUsed,
		extraction.ProcessedAt, extraction.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create extraction: %w", err)
	}
	return nil
}

func (r *ExtractionRepo) DeleteByEmailID(ctx context.Context, emailID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM extraction WHERE email_id = $1`, emailID)
	if err != nil {
		return fmt.Errorf("delete extraction by email: %w", err)
	}
	return nil
}

func (r *ExtractionRepo) DeleteByEmailIDScoped(ctx context.Context, userID, emailID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM extraction WHERE email_id = $1
		 AND EXISTS (SELECT 1 FROM email WHERE email.id = $1 AND email.user_id = $2)`,
		emailID, userID)
	if err != nil {
		return fmt.Errorf("delete extraction by email (scoped): %w", err)
	}
	return nil
}

func (r *ExtractionRepo) GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email_id, summary, key_points, topics, links,
		    sentiment, source_category, provider, model, tokens_used, processed_at, created_at
		 FROM extraction WHERE email_id = $1`, emailID)

	return scanExtraction(row)
}

func (r *ExtractionRepo) GetByEmailIDScoped(ctx context.Context, userID, emailID string) (*domain.Extraction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT e.id, e.email_id, e.summary, e.key_points, e.topics, e.links,
		    e.sentiment, e.source_category, e.provider, e.model, e.tokens_used, e.processed_at, e.created_at
		 FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 WHERE e.email_id = $1 AND em.user_id = $2`, emailID, userID)

	return scanExtraction(row)
}

func (r *ExtractionRepo) ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time, excludedCategories ...string) ([]domain.Extraction, error) {
	if len(excludedCategories) == 0 {
		excludedCategories = []string{"notification", "transactional", "marketing"}
	}

	query := `SELECT e.id, e.email_id, e.summary, e.key_points, e.topics, e.links,
		    e.sentiment, e.source_category, e.provider, e.model, e.tokens_used, e.processed_at, e.created_at
		 FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 WHERE em.user_id = $1 AND em.received_at >= $2 AND em.received_at < $3
		   AND em.status = 'processed'`

	args := []any{userID, from, to}
	if len(excludedCategories) > 0 {
		query += ` AND e.source_category != ALL($4)`
		args = append(args, excludedCategories)
	}
	query += ` ORDER BY em.received_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list extractions: %w", err)
	}
	defer rows.Close()

	var extractions []domain.Extraction
	for rows.Next() {
		ext, err := scanExtraction(rows)
		if err != nil {
			return nil, err
		}
		extractions = append(extractions, *ext)
	}
	return extractions, nil
}

func (r *ExtractionRepo) GetSummariesByEmailIDs(ctx context.Context, emailIDs []string) (map[string]string, error) {
	if len(emailIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT email_id, summary FROM extraction WHERE email_id = ANY($1)`, emailIDs)
	if err != nil {
		return nil, fmt.Errorf("get summaries by email IDs: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string, len(emailIDs))
	for rows.Next() {
		var emailID, summary string
		if err := rows.Scan(&emailID, &summary); err != nil {
			return nil, fmt.Errorf("scan summary: %w", err)
		}
		result[emailID] = summary
	}
	return result, nil
}

func (r *ExtractionRepo) ListTopicsByUser(ctx context.Context, userID string, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT topic FROM (
		   SELECT topic, COUNT(*) AS cnt FROM extraction e
		   JOIN email em ON em.id = e.email_id
		   CROSS JOIN LATERAL jsonb_array_elements_text(e.topics) AS topic
		   WHERE em.user_id = $1
		   GROUP BY topic ORDER BY cnt DESC LIMIT $2
		 ) t ORDER BY topic`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	defer rows.Close()

	var topics []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan topic: %w", err)
		}
		topics = append(topics, t)
	}
	return topics, nil
}

func (r *ExtractionRepo) ListTopicsWithCount(ctx context.Context, userID string, limit int) ([]domain.TopicCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT topic, COUNT(*) AS cnt FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 CROSS JOIN LATERAL jsonb_array_elements_text(e.topics) AS topic
		 WHERE em.user_id = $1
		 GROUP BY topic ORDER BY cnt DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list topics with count: %w", err)
	}
	defer rows.Close()

	var topics []domain.TopicCount
	for rows.Next() {
		var tc domain.TopicCount
		var cnt int64
		if err := rows.Scan(&tc.Topic, &cnt); err != nil {
			return nil, fmt.Errorf("scan topic count: %w", err)
		}
		tc.Count = int(cnt)
		topics = append(topics, tc)
	}
	return topics, nil
}

func (r *ExtractionRepo) CountByCategory(ctx context.Context, userID string) ([]domain.CategoryCount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT COALESCE(e.source_category, 'other') AS cat, COUNT(*) AS cnt
		 FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 WHERE em.user_id = $1
		 GROUP BY cat ORDER BY cnt DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("count by category: %w", err)
	}
	defer rows.Close()

	var counts []domain.CategoryCount
	for rows.Next() {
		var cc domain.CategoryCount
		var cnt int64
		if err := rows.Scan(&cc.Category, &cnt); err != nil {
			return nil, fmt.Errorf("scan category count: %w", err)
		}
		cc.Count = int(cnt)
		counts = append(counts, cc)
	}
	return counts, nil
}

func scanExtraction(row scannable) (*domain.Extraction, error) {
	var e domain.Extraction
	var keyPoints, topics, links []byte

	err := row.Scan(
		&e.ID, &e.EmailID, &e.Summary, &keyPoints, &topics, &links,
		&e.Sentiment, &e.SourceCategory, &e.Provider, &e.Model, &e.TokensUsed,
		&e.ProcessedAt, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan extraction: %w", err)
	}

	json.Unmarshal(keyPoints, &e.KeyPoints)
	json.Unmarshal(topics, &e.Topics)
	json.Unmarshal(links, &e.Links)

	return &e, nil
}
