package database

import (
	"context"
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
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&totalUsers)
	stats.TotalUsers = int(totalUsers)

	// Active users this week (users who received emails)
	var activeUsersWeek int64
	r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&activeUsersWeek)
	stats.ActiveUsersWeek = int(activeUsersWeek)

	// Total emails
	var totalEmails int64
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM email`).Scan(&totalEmails)
	stats.TotalEmails = int(totalEmails)

	// Emails this week
	var emailsThisWeek int64
	r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&emailsThisWeek)
	stats.EmailsThisWeek = int(emailsThisWeek)

	// Total digests
	var totalDigests int64
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM digest`).Scan(&totalDigests)
	stats.TotalDigests = int(totalDigests)

	// Digests this week
	var digestsThisWeek int64
	r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM digest WHERE generated_at >= $1`, weekAgo,
	).Scan(&digestsThisWeek)
	stats.DigestsThisWeek = int(digestsThisWeek)

	// Top 5 global senders this week
	rows, err := r.pool.Query(ctx,
		`SELECT from_address, from_name, COUNT(*) as cnt
		 FROM email WHERE received_at >= $1
		 GROUP BY from_address, from_name ORDER BY cnt DESC LIMIT 5`, weekAgo)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s domain.SenderCount
			var cnt int64
			if err := rows.Scan(&s.FromAddress, &s.FromName, &cnt); err == nil {
				s.Count = int(cnt)
				stats.TopGlobalSenders = append(stats.TopGlobalSenders, s)
			}
		}
	}
	if stats.TopGlobalSenders == nil {
		stats.TopGlobalSenders = []domain.SenderCount{}
	}

	// Processing status breakdown (all time)
	statusRows, err := r.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM email GROUP BY status`)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int64
			if err := statusRows.Scan(&status, &count); err == nil {
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
		}
	}

	// Feedback summary (all time)
	var feedbackUseful, feedbackNotUseful int64
	r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE rating = 'useful'), COUNT(*) FILTER (WHERE rating = 'not_useful') FROM email_feedback`,
	).Scan(&feedbackUseful, &feedbackNotUseful)
	stats.FeedbackUseful = int(feedbackUseful)
	stats.FeedbackNotUseful = int(feedbackNotUseful)

	// Weekly email trend (last 8 weeks)
	weeklyEmailRows, err := r.pool.Query(ctx,
		`SELECT DATE_TRUNC('week', received_at) AS week, COUNT(*)
		 FROM email WHERE received_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
	if err != nil {
		slog.Error("admin stats: weekly emails query failed", "error", err)
	} else {
		defer weeklyEmailRows.Close()
		for weeklyEmailRows.Next() {
			var wc domain.WeekCount
			var count int64
			if err := weeklyEmailRows.Scan(&wc.Week, &count); err != nil {
				slog.Error("admin stats: weekly emails scan failed", "error", err)
			} else {
				wc.Count = int(count)
				stats.WeeklyEmails = append(stats.WeeklyEmails, wc)
			}
		}
	}
	slog.Info("admin stats: weekly emails", "count", len(stats.WeeklyEmails))
	if stats.WeeklyEmails == nil {
		stats.WeeklyEmails = []domain.WeekCount{}
	}

	// Weekly digest trend (last 8 weeks)
	weeklyDigestRows, err := r.pool.Query(ctx,
		`SELECT DATE_TRUNC('week', generated_at) AS week, COUNT(*)
		 FROM digest WHERE generated_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
	if err != nil {
		slog.Error("admin stats: weekly digests query failed", "error", err)
	} else {
		defer weeklyDigestRows.Close()
		for weeklyDigestRows.Next() {
			var wc domain.WeekCount
			var count int64
			if err := weeklyDigestRows.Scan(&wc.Week, &count); err != nil {
				slog.Error("admin stats: weekly digests scan failed", "error", err)
			} else {
				wc.Count = int(count)
				stats.WeeklyDigests = append(stats.WeeklyDigests, wc)
			}
		}
	}
	if stats.WeeklyDigests == nil {
		stats.WeeklyDigests = []domain.WeekCount{}
	}

	// Weekly token usage trend (last 8 weeks)
	weeklyTokenRows, err := r.pool.Query(ctx,
		`SELECT DATE_TRUNC('week', created_at) AS week, COALESCE(SUM(total_tokens), 0)
		 FROM token_usage WHERE created_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
	if err != nil {
		slog.Error("admin stats: weekly tokens query failed", "error", err)
	}
	if err == nil {
		defer weeklyTokenRows.Close()
		for weeklyTokenRows.Next() {
			var wc domain.WeekCount
			var tokens int64
			if err := weeklyTokenRows.Scan(&wc.Week, &tokens); err == nil {
				wc.Count = int(tokens)
				stats.WeeklyTokens = append(stats.WeeklyTokens, wc)
			}
		}
	}
	if stats.WeeklyTokens == nil {
		stats.WeeklyTokens = []domain.WeekCount{}
	}

	return &stats, nil
}

// ListUsers returns all users with their email counts and token usage for admin display.
func (r *AdminStatsRepo) ListUsers(ctx context.Context) ([]domain.AdminUser, error) {
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

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
		ORDER BY COALESCE(tu.tokens, 0) DESC`, monthStart)
	if err != nil {
		return nil, err
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

	return users, nil
}
