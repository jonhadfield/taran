package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminStatsRepo provides admin dashboard queries that span multiple tables.
type AdminStatsRepo struct {
	pool *pgxpool.Pool
}

func NewAdminStatsRepo(pool *pgxpool.Pool) *AdminStatsRepo {
	return &AdminStatsRepo{pool: pool}
}

// GetStats returns aggregate statistics for the admin dashboard.
func (r *AdminStatsRepo) GetStats(ctx context.Context) (*domain.AdminStats, error) {
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	var stats domain.AdminStats

	// Total users
	var totalUsers int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&totalUsers); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	stats.TotalUsers = int(totalUsers)

	// Active users this week (users who received emails)
	var activeUsersWeek int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&activeUsersWeek); err != nil {
		return nil, fmt.Errorf("count active users: %w", err)
	}
	stats.ActiveUsersWeek = int(activeUsersWeek)

	// Total emails
	var totalEmails int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM email`).Scan(&totalEmails); err != nil {
		return nil, fmt.Errorf("count emails: %w", err)
	}
	stats.TotalEmails = int(totalEmails)

	// Emails this week
	var emailsThisWeek int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&emailsThisWeek); err != nil {
		return nil, fmt.Errorf("count emails this week: %w", err)
	}
	stats.EmailsThisWeek = int(emailsThisWeek)

	// Total digests
	var totalDigests int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM digest`).Scan(&totalDigests); err != nil {
		return nil, fmt.Errorf("count digests: %w", err)
	}
	stats.TotalDigests = int(totalDigests)

	// Digests this week
	var digestsThisWeek int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM digest WHERE generated_at >= $1`, weekAgo,
	).Scan(&digestsThisWeek); err != nil {
		return nil, fmt.Errorf("count digests this week: %w", err)
	}
	stats.DigestsThisWeek = int(digestsThisWeek)

	// Top 5 global senders this week
	topSenders, err := scanSenderCounts(ctx, r.pool, weekAgo)
	if err != nil {
		return nil, err
	}
	stats.TopGlobalSenders = topSenders

	// Processing status breakdown (all time)
	if err := scanStatusCounts(ctx, r.pool, &stats); err != nil {
		return nil, err
	}

	// Feedback summary (all time)
	var feedbackUseful, feedbackNotUseful int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE rating = 'useful'), COUNT(*) FILTER (WHERE rating = 'not_useful') FROM email_feedback`,
	).Scan(&feedbackUseful, &feedbackNotUseful); err != nil {
		return nil, fmt.Errorf("count feedback: %w", err)
	}
	stats.FeedbackUseful = int(feedbackUseful)
	stats.FeedbackNotUseful = int(feedbackNotUseful)

	// Weekly email trend (last 8 weeks)
	weeklyEmails, err := scanWeekCounts(ctx, r.pool,
		`SELECT DATE_TRUNC('week', received_at) AS week, COUNT(*)
		 FROM email WHERE received_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`,
		"weekly emails")
	if err != nil {
		return nil, err
	}
	stats.WeeklyEmails = weeklyEmails

	// Weekly digest trend (last 8 weeks)
	weeklyDigests, err := scanWeekCounts(ctx, r.pool,
		`SELECT DATE_TRUNC('week', generated_at) AS week, COUNT(*)
		 FROM digest WHERE generated_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`,
		"weekly digests")
	if err != nil {
		return nil, err
	}
	stats.WeeklyDigests = weeklyDigests

	// Weekly token usage trend (last 8 weeks)
	weeklyTokens, err := scanWeekCounts(ctx, r.pool,
		`SELECT DATE_TRUNC('week', created_at) AS week, COALESCE(SUM(total_tokens), 0)
		 FROM token_usage WHERE created_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`,
		"weekly tokens")
	if err != nil {
		return nil, err
	}
	stats.WeeklyTokens = weeklyTokens

	return &stats, nil
}

func scanSenderCounts(ctx context.Context, pool *pgxpool.Pool, weekAgo time.Time) ([]domain.SenderCount, error) {
	rows, err := pool.Query(ctx,
		`SELECT from_address, from_name, COUNT(*) as cnt
		 FROM email WHERE received_at >= $1
		 GROUP BY from_address, from_name ORDER BY cnt DESC LIMIT 5`, weekAgo)
	if err != nil {
		return nil, fmt.Errorf("top global senders: %w", err)
	}
	defer rows.Close()

	senders := []domain.SenderCount{}
	for rows.Next() {
		var s domain.SenderCount
		var cnt int64
		if err := rows.Scan(&s.FromAddress, &s.FromName, &cnt); err != nil {
			return nil, fmt.Errorf("scan top sender: %w", err)
		}
		s.Count = int(cnt)
		senders = append(senders, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top senders: %w", err)
	}
	return senders, nil
}

func scanStatusCounts(ctx context.Context, pool *pgxpool.Pool, stats *domain.AdminStats) error {
	rows, err := pool.Query(ctx, `SELECT status, COUNT(*) FROM email GROUP BY status`)
	if err != nil {
		return fmt.Errorf("status breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return fmt.Errorf("scan status count: %w", err)
		}
		switch status {
		case "processed":
			stats.ProcessedCount = int(count)
		case "failed":
			stats.FailedCount = int(count)
		case "skipped":
			stats.SkippedCount = int(count)
		case "pending", "processing":
			stats.PendingCount += int(count)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate status counts: %w", err)
	}
	return nil
}

func scanWeekCounts(ctx context.Context, pool *pgxpool.Pool, query, label string) ([]domain.WeekCount, error) {
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s query: %w", label, err)
	}
	defer rows.Close()

	counts := []domain.WeekCount{}
	for rows.Next() {
		var wc domain.WeekCount
		var count int64
		if err := rows.Scan(&wc.Week, &count); err != nil {
			return nil, fmt.Errorf("scan %s: %w", label, err)
		}
		wc.Count = int(count)
		counts = append(counts, wc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", label, err)
	}
	return counts, nil
}

// ListUsers returns all users with their email counts and token usage for admin display.
// ListUsers returns one page of users, heaviest token users first, along with
// the total number of users so the caller can page through them.
func (r *AdminStatsRepo) ListUsers(ctx context.Context, limit, offset int) ([]domain.AdminUser, int, error) {
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			u.id,
			u.email,
			COALESCE(u.name, '') as name,
			COALESCE(ec.cnt, 0) as email_count,
			COALESCE(tu.tokens, 0) as monthly_tokens,
			COALESCE(p.monthly_token_limit, 0) as token_limit
		FROM "user" u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as cnt FROM email GROUP BY user_id
		) ec ON ec.user_id = u.id
		LEFT JOIN (
			SELECT user_id, SUM(total_tokens) as tokens
			FROM token_usage WHERE created_at >= $1
			GROUP BY user_id
		) tu ON tu.user_id = u.id
		LEFT JOIN user_preference p ON p.user_id = u.id
		ORDER BY COALESCE(tu.tokens, 0) DESC, u.id
		LIMIT $2 OFFSET $3`, monthStart, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []domain.AdminUser
	for rows.Next() {
		var u domain.AdminUser
		var emailCount, monthlyTokens, tokenLimit int64
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &emailCount, &monthlyTokens, &tokenLimit); err != nil {
			slog.Error("admin: failed to scan user row", "error", err)
			continue
		}
		u.EmailCount = int(emailCount)
		u.MonthlyTokensUsed = int(monthlyTokens)
		u.MonthlyTokenLimit = int(tokenLimit)
		users = append(users, u)
	}
	if users == nil {
		users = []domain.AdminUser{}
	}

	return users, int(total), nil
}
