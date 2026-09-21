package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalysisRuleRepo struct {
	pool *pgxpool.Pool
}

func NewAnalysisRuleRepo(pool *pgxpool.Pool) *AnalysisRuleRepo {
	return &AnalysisRuleRepo{pool: pool}
}

func (r *AnalysisRuleRepo) Create(ctx context.Context, rule *domain.AnalysisRule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO analysis_rule (id, user_id, rule, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		rule.ID, rule.UserID, rule.Rule, rule.IsActive)
	if err != nil {
		return fmt.Errorf("create analysis rule: %w", err)
	}
	return nil
}

// ListByUser returns all of a user's rules, oldest first, so that the order
// shown in the UI matches the order the rules appear in LLM prompts.
func (r *AnalysisRuleRepo) ListByUser(ctx context.Context, userID string) ([]domain.AnalysisRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, rule, is_active, created_at, updated_at
		 FROM analysis_rule WHERE user_id = $1 ORDER BY created_at ASC, id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list analysis rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.AnalysisRule
	for rows.Next() {
		var a domain.AnalysisRule
		if err := rows.Scan(&a.ID, &a.UserID, &a.Rule, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan analysis rule: %w", err)
		}
		rules = append(rules, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analysis rules: %w", err)
	}
	return rules, nil
}

// ListActiveRules returns the text of a user's active rules, oldest first.
func (r *AnalysisRuleRepo) ListActiveRules(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT rule FROM analysis_rule
		 WHERE user_id = $1 AND is_active ORDER BY created_at ASC, id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list active analysis rules: %w", err)
	}
	defer rows.Close()

	var rules []string
	for rows.Next() {
		var rule string
		if err := rows.Scan(&rule); err != nil {
			return nil, fmt.Errorf("scan analysis rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analysis rules: %w", err)
	}
	return rules, nil
}

// Update changes the text and/or active flag of a rule. Nil fields are left
// unchanged. Returns ErrNotFound if the rule does not exist for this user.
func (r *AnalysisRuleRepo) Update(ctx context.Context, userID, id string, rule *string, isActive *bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE analysis_rule
		 SET rule = COALESCE($3, rule), is_active = COALESCE($4, is_active), updated_at = NOW()
		 WHERE id = $1 AND user_id = $2`, id, userID, rule, isActive)
	if err != nil {
		return fmt.Errorf("update analysis rule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AnalysisRuleRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM analysis_rule WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete analysis rule: %w", err)
	}
	return nil
}

func (r *AnalysisRuleRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM analysis_rule WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count analysis rules: %w", err)
	}
	return int(count), nil
}
