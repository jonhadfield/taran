package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminStatsHandler struct {
	Pool        *pgxpool.Pool
	LLMProvider string
	LLMModel    string
	TokenUsage  database.TokenUsageRepository
	Preferences database.PreferenceRepository
	AppSettings *database.AppSettingRepo
}

func (h *AdminStatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	var stats domain.AdminStats

	// Total users
	h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&stats.TotalUsers)

	// Active users this week (users who received emails)
	h.Pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&stats.ActiveUsersWeek)

	// Total emails
	h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM email`).Scan(&stats.TotalEmails)

	// Emails this week
	h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM email WHERE received_at >= $1`, weekAgo,
	).Scan(&stats.EmailsThisWeek)

	// Total digests
	h.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM digest`).Scan(&stats.TotalDigests)

	// Digests this week
	h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM digest WHERE generated_at >= $1`, weekAgo,
	).Scan(&stats.DigestsThisWeek)

	// Top 5 global senders this week
	rows, err := h.Pool.Query(ctx,
		`SELECT from_address, from_name, COUNT(*) as cnt
		 FROM email WHERE received_at >= $1
		 GROUP BY from_address, from_name ORDER BY cnt DESC LIMIT 5`, weekAgo)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s domain.SenderCount
			if err := rows.Scan(&s.FromAddress, &s.FromName, &s.Count); err == nil {
				stats.TopGlobalSenders = append(stats.TopGlobalSenders, s)
			}
		}
	}
	if stats.TopGlobalSenders == nil {
		stats.TopGlobalSenders = []domain.SenderCount{}
	}

	// Processing status breakdown (all time)
	statusRows, err := h.Pool.Query(ctx,
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
	h.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE rating = 'useful'), COUNT(*) FILTER (WHERE rating = 'not_useful') FROM email_feedback`,
	).Scan(&stats.FeedbackUseful, &stats.FeedbackNotUseful)

	// Weekly email trend (last 8 weeks)
	weeklyEmailRows, err := h.Pool.Query(ctx,
		`SELECT DATE_TRUNC('week', received_at) AS week, COUNT(*)
		 FROM email WHERE received_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
	if err == nil {
		defer weeklyEmailRows.Close()
		for weeklyEmailRows.Next() {
			var wc domain.WeekCount
			if err := weeklyEmailRows.Scan(&wc.Week, &wc.Count); err == nil {
				stats.WeeklyEmails = append(stats.WeeklyEmails, wc)
			}
		}
	}
	if stats.WeeklyEmails == nil {
		stats.WeeklyEmails = []domain.WeekCount{}
	}

	// Weekly digest trend (last 8 weeks)
	weeklyDigestRows, err := h.Pool.Query(ctx,
		`SELECT DATE_TRUNC('week', generated_at) AS week, COUNT(*)
		 FROM digest WHERE generated_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
	if err == nil {
		defer weeklyDigestRows.Close()
		for weeklyDigestRows.Next() {
			var wc domain.WeekCount
			if err := weeklyDigestRows.Scan(&wc.Week, &wc.Count); err == nil {
				stats.WeeklyDigests = append(stats.WeeklyDigests, wc)
			}
		}
	}
	if stats.WeeklyDigests == nil {
		stats.WeeklyDigests = []domain.WeekCount{}
	}

	// Weekly token usage trend (last 8 weeks)
	weeklyTokenRows, err := h.Pool.Query(ctx,
		`SELECT DATE_TRUNC('week', created_at) AS week, COALESCE(SUM(total_tokens), 0)
		 FROM token_usage WHERE created_at >= NOW() - INTERVAL '8 weeks'
		 GROUP BY week ORDER BY week`)
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

	stats.LLMProvider = h.LLMProvider
	stats.LLMModel = h.LLMModel

	if h.AppSettings != nil {
		limit, _ := h.AppSettings.GetInt(ctx, "default_monthly_token_limit", 500000)
		stats.DefaultMonthlyTokenLimit = limit
	}

	if h.TokenUsage != nil {
		total, err := h.TokenUsage.GetGlobalMonthlyTotal(ctx)
		if err == nil {
			stats.MonthlyTokensUsed = total
		}
	}

	WriteJSON(w, http.StatusOK, stats)
}

func (h *AdminStatsHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := h.Pool.Query(ctx, `
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
		slog.Error("admin: failed to query users", "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	defaultLimit := 500000
	if h.AppSettings != nil {
		if v, err := h.AppSettings.GetInt(ctx, "default_monthly_token_limit", 500000); err == nil {
			defaultLimit = v
		}
	}

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
		// If user has no custom limit, show the global default
		if u.MonthlyTokenLimit == 0 {
			u.MonthlyTokenLimit = defaultLimit
		}
		users = append(users, u)
	}
	if users == nil {
		users = []domain.AdminUser{}
	}

	WriteJSON(w, http.StatusOK, users)
}

func (h *AdminStatsHandler) SetDefaultTokenLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DefaultMonthlyTokenLimit int `json:"DefaultMonthlyTokenLimit"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DefaultMonthlyTokenLimit < 0 {
		WriteError(w, http.StatusBadRequest, "token limit cannot be negative")
		return
	}

	if err := h.AppSettings.Set(r.Context(), "default_monthly_token_limit", fmt.Sprintf("%d", req.DefaultMonthlyTokenLimit)); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update default token limit")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":                   "updated",
		"DefaultMonthlyTokenLimit": req.DefaultMonthlyTokenLimit,
	})
}

func (h *AdminStatsHandler) SetUserTokenLimit(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		WriteError(w, http.StatusBadRequest, "user id required")
		return
	}

	var req struct {
		MonthlyTokenLimit int `json:"MonthlyTokenLimit"`
	}
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.MonthlyTokenLimit < 0 {
		WriteError(w, http.StatusBadRequest, "token limit cannot be negative")
		return
	}

	pref, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get user preferences")
		return
	}

	pref.MonthlyTokenLimit = req.MonthlyTokenLimit
	if err := h.Preferences.Upsert(r.Context(), pref); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update token limit")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"status":            "updated",
		"UserID":            userID,
		"MonthlyTokenLimit": req.MonthlyTokenLimit,
	})
}
