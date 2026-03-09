package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenUsageRepo struct {
	pool *pgxpool.Pool
}

func NewTokenUsageRepo(pool *pgxpool.Pool) *TokenUsageRepo {
	return &TokenUsageRepo{pool: pool}
}

func (r *TokenUsageRepo) Create(ctx context.Context, tu *domain.TokenUsage) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO token_usage (id, user_id, operation, provider, model,
		    input_tokens, output_tokens, total_tokens, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		tu.ID, tu.UserID, tu.Operation, tu.Provider, tu.Model,
		tu.InputTokens, tu.OutputTokens, tu.TotalTokens, tu.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create token usage: %w", err)
	}
	return nil
}

func (r *TokenUsageRepo) GetMonthlyTotal(ctx context.Context, userID string) (int, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_tokens), 0) FROM token_usage
		 WHERE user_id = $1 AND created_at >= $2`,
		userID, monthStart,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get monthly total: %w", err)
	}
	return total, nil
}

func (r *TokenUsageRepo) GetUsageStats(ctx context.Context, userID string) (*domain.UsageStats, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	stats := &domain.UsageStats{
		PeriodStart: monthStart,
		PeriodEnd:   monthStart.AddDate(0, 1, 0),
	}

	// Monthly totals by operation
	rows, err := r.pool.Query(ctx,
		`SELECT operation, COALESCE(SUM(total_tokens), 0)
		 FROM token_usage
		 WHERE user_id = $1 AND created_at >= $2
		 GROUP BY operation`,
		userID, monthStart,
	)
	if err != nil {
		return nil, fmt.Errorf("get usage stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var op string
		var tokens int
		if err := rows.Scan(&op, &tokens); err != nil {
			return nil, fmt.Errorf("scan usage stats: %w", err)
		}
		stats.MonthlyTokensUsed += tokens
		switch op {
		case "triage":
			stats.TriageTokens = tokens
		case "extract":
			stats.ExtractTokens = tokens
		case "digest":
			stats.DigestTokens = tokens
		}
	}

	// Daily total
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_tokens), 0) FROM token_usage
		 WHERE user_id = $1 AND created_at >= $2`,
		userID, dayStart,
	).Scan(&stats.DailyTokensUsed)
	if err != nil {
		return nil, fmt.Errorf("get daily usage: %w", err)
	}

	return stats, nil
}

func (r *TokenUsageRepo) GetGlobalMonthlyTotal(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_tokens), 0) FROM token_usage WHERE created_at >= $1`,
		monthStart,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get global monthly total: %w", err)
	}
	return total, nil
}
