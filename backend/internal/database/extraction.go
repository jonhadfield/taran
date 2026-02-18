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
	actionItems, _ := json.Marshal(extraction.ActionItems)

	_, err := r.pool.Exec(ctx,
		`INSERT INTO extraction (id, email_id, summary, key_points, topics, links, action_items,
		    sentiment, source_category, provider, model, tokens_used, processed_at, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		extraction.ID, extraction.EmailID, extraction.Summary,
		keyPoints, topics, links, actionItems,
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

func (r *ExtractionRepo) GetByEmailID(ctx context.Context, emailID string) (*domain.Extraction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, email_id, summary, key_points, topics, links, action_items,
		    sentiment, source_category, provider, model, tokens_used, processed_at, created_at
		 FROM extraction WHERE email_id = $1`, emailID)

	return scanExtraction(row)
}

func (r *ExtractionRepo) ListByUserAndPeriod(ctx context.Context, userID string, from, to time.Time) ([]domain.Extraction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.email_id, e.summary, e.key_points, e.topics, e.links, e.action_items,
		    e.sentiment, e.source_category, e.provider, e.model, e.tokens_used, e.processed_at, e.created_at
		 FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 WHERE em.user_id = $1 AND em.received_at >= $2 AND em.received_at < $3
		   AND em.status = 'processed'
		   AND e.source_category NOT IN ('notification', 'transactional', 'marketing')
		 ORDER BY em.received_at DESC`, userID, from, to)
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

func (r *ExtractionRepo) ListTopicsByUser(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT topic FROM extraction e
		 JOIN email em ON em.id = e.email_id
		 CROSS JOIN LATERAL jsonb_array_elements_text(e.topics) AS topic
		 WHERE em.user_id = $1 ORDER BY topic`, userID)
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

func scanExtraction(row scannable) (*domain.Extraction, error) {
	var e domain.Extraction
	var keyPoints, topics, links, actionItems []byte

	err := row.Scan(
		&e.ID, &e.EmailID, &e.Summary, &keyPoints, &topics, &links, &actionItems,
		&e.Sentiment, &e.SourceCategory, &e.Provider, &e.Model, &e.TokensUsed,
		&e.ProcessedAt, &e.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan extraction: %w", err)
	}

	json.Unmarshal(keyPoints, &e.KeyPoints)
	json.Unmarshal(topics, &e.Topics)
	json.Unmarshal(links, &e.Links)
	json.Unmarshal(actionItems, &e.ActionItems)

	return &e, nil
}
