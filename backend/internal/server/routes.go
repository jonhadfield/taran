package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouterDeps struct {
	Pool               *pgxpool.Pool
	WebhookSecret      string
	APIKey             string
	WebhookHandler     *handler.WebhookHandler
	EmailHandler       *handler.EmailHandler
	DigestHandler      *handler.DigestHandler
	AccountHandler     *handler.AccountHandler
	PreferenceHandler  *handler.PreferenceHandler
	SenderHandler      *handler.SenderHandler
	StatsHandler       *handler.StatsHandler
	StatsHistoryHandler *handler.StatsHistoryHandler
	TopicHandler       *handler.TopicHandler
	FeedbackHandler    *handler.FeedbackHandler
	AdminStatsHandler  *handler.AdminStatsHandler
	DashboardHandler   *handler.DashboardHandler
	InviteHandler      *handler.InviteHandler
	SessionAuth        *auth.SessionAuth
}

func NewRouter(deps RouterDeps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth(deps.Pool))

	// Webhook (shared secret auth)
	webhookAuth := auth.WebhookAuth(deps.WebhookSecret, http.HandlerFunc(deps.WebhookHandler.IngestEmail))
	mux.Handle("POST /webhook/email", webhookAuth)

	// API routes (session auth) — public endpoints registered here but skipped by auth
	api := http.NewServeMux()

	// Public endpoints (no auth — skipped by publicPathAuth wrapper below)
	api.HandleFunc("GET /api/public/digests/{token}", deps.DigestHandler.GetPublic)
	api.HandleFunc("POST /api/public/unsubscribe", deps.PreferenceHandler.Unsubscribe)
	api.HandleFunc("GET /api/dashboard", deps.DashboardHandler.Get)
	api.HandleFunc("GET /api/emails", deps.EmailHandler.List)
	api.HandleFunc("GET /api/emails/{id}", deps.EmailHandler.Get)
	api.HandleFunc("PATCH /api/emails/{id}", deps.EmailHandler.UpdateState)
	api.HandleFunc("DELETE /api/emails/{id}", deps.EmailHandler.Delete)
	api.HandleFunc("POST /api/emails/{id}/reprocess", deps.EmailHandler.Reprocess)
	api.HandleFunc("GET /api/digests", deps.DigestHandler.List)
	api.HandleFunc("GET /api/digests/{id}", deps.DigestHandler.Get)
	api.HandleFunc("DELETE /api/digests/{id}", deps.DigestHandler.Delete)
	api.HandleFunc("POST /api/digests/{id}/share", deps.DigestHandler.Share)
	api.HandleFunc("DELETE /api/digests/{id}/share", deps.DigestHandler.Unshare)
	api.HandleFunc("GET /api/accounts", deps.AccountHandler.List)
	api.HandleFunc("POST /api/accounts", deps.AccountHandler.Create)
	api.HandleFunc("DELETE /api/accounts/{id}", deps.AccountHandler.Delete)
	api.HandleFunc("GET /api/accounts/check-username", deps.AccountHandler.CheckUsername)
	api.HandleFunc("GET /api/preferences", deps.PreferenceHandler.Get)
	api.HandleFunc("PATCH /api/preferences", deps.PreferenceHandler.Update)
	api.HandleFunc("GET /api/senders", deps.SenderHandler.List)
	api.HandleFunc("GET /api/senders/suggestions", deps.SenderHandler.Suggestions)
	api.HandleFunc("PATCH /api/senders", deps.SenderHandler.Update)
	api.HandleFunc("GET /api/stats", deps.StatsHandler.Get)
	api.HandleFunc("GET /api/stats/history", deps.StatsHistoryHandler.Get)
	api.HandleFunc("POST /api/digests/generate", deps.DigestHandler.Generate)
	api.HandleFunc("GET /api/topics", deps.TopicHandler.List)
	api.HandleFunc("POST /api/emails/{id}/feedback", deps.FeedbackHandler.Upsert)
	api.HandleFunc("GET /api/emails/{id}/feedback", deps.FeedbackHandler.Get)
	api.HandleFunc("DELETE /api/emails/{id}/feedback", deps.FeedbackHandler.Delete)
	api.HandleFunc("GET /api/access", deps.InviteHandler.CheckAccess)

	// Admin routes
	admin := http.NewServeMux()
	admin.HandleFunc("POST /api/admin/digests/{id}/send", deps.DigestHandler.SendEmail)
	admin.HandleFunc("POST /api/admin/digests/generate", deps.DigestHandler.Generate)
	admin.HandleFunc("GET /api/admin/stats", deps.AdminStatsHandler.Get)
	admin.HandleFunc("POST /api/admin/invites", deps.InviteHandler.Create)
	admin.HandleFunc("GET /api/admin/invites", deps.InviteHandler.List)
	api.Handle("/api/admin/", deps.SessionAuth.AdminOnly(admin))

	authedAPI := publicPathSkip(auth.APIKeyAuth(deps.APIKey, deps.SessionAuth.Middleware(api)), api)
	mux.Handle("/api/", authedAPI)

	return mux
}

// publicPathSkip bypasses auth for /api/public/ paths, forwarding directly to the handler.
func publicPathSkip(authed, raw http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/public/") {
			raw.ServeHTTP(w, r)
			return
		}
		authed.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds standard security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func handleHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if pool != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if err := pool.Ping(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "reason": "database unreachable"})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
