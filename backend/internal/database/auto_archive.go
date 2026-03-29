package database

import (
	"context"
	"fmt"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AutoArchiveRuleRepo struct {
	pool *pgxpool.Pool
}

func NewAutoArchiveRuleRepo(pool *pgxpool.Pool) *AutoArchiveRuleRepo {
	return &AutoArchiveRuleRepo{pool: pool}
}

func (r *AutoArchiveRuleRepo) ListByUser(ctx context.Context, userID string) ([]domain.AutoArchiveRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, rule_type, rule_value, archive_after_days, is_active, created_at, updated_at
		 FROM auto_archive_rule WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list auto-archive rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.AutoArchiveRule
	for rows.Next() {
		var rule domain.AutoArchiveRule
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.RuleType, &rule.RuleValue, &rule.ArchiveAfterDays, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan auto-archive rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *AutoArchiveRuleRepo) Upsert(ctx context.Context, rule *domain.AutoArchiveRule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO auto_archive_rule (id, user_id, rule_type, rule_value, archive_after_days, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		 ON CONFLICT (user_id, rule_type, rule_value) DO UPDATE SET
		     archive_after_days = $5, is_active = $6, updated_at = NOW()`,
		rule.ID, rule.UserID, rule.RuleType, rule.RuleValue, rule.ArchiveAfterDays, rule.IsActive)
	if err != nil {
		return fmt.Errorf("upsert auto-archive rule: %w", err)
	}
	return nil
}

func (r *AutoArchiveRuleRepo) Delete(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM auto_archive_rule WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete auto-archive rule: %w", err)
	}
	return nil
}

func (r *AutoArchiveRuleRepo) ListEmailsToArchive(ctx context.Context, limit int) ([]domain.ArchiveCandidate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT e.id, e.user_id
		 FROM auto_archive_rule r
		 JOIN email e ON e.user_id = r.user_id AND e.is_archived = false AND e.status = 'processed'
		 LEFT JOIN extraction ex ON ex.email_id = e.id
		 LEFT JOIN sender_preference sp ON sp.user_id = e.user_id AND sp.from_address = e.from_address
		 WHERE r.is_active = true
		   AND (
		     (r.rule_type = 'sender' AND e.from_address = r.rule_value)
		     OR (r.rule_type = 'category' AND (
		         COALESCE(NULLIF(sp.category, ''), ex.source_category, '') = r.rule_value
		     ))
		   )
		   AND e.received_at < NOW() - (r.archive_after_days || ' days')::INTERVAL
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list emails to archive: %w", err)
	}
	defer rows.Close()

	var candidates []domain.ArchiveCandidate
	for rows.Next() {
		var c domain.ArchiveCandidate
		if err := rows.Scan(&c.EmailID, &c.UserID); err != nil {
			return nil, fmt.Errorf("scan archive candidate: %w", err)
		}
		candidates = append(candidates, c)
	}
	return candidates, nil
}
