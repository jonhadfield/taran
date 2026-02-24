package handler

import (
	"net/http"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminStatsHandler struct {
	Pool        *pgxpool.Pool
	LLMProvider string
	LLMModel    string
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

	stats.LLMProvider = h.LLMProvider
	stats.LLMModel = h.LLMModel

	WriteJSON(w, http.StatusOK, stats)
}
