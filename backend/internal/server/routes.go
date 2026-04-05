package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/handler"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RouterDeps struct {
	Pool               *pgxpool.Pool
	WebhookSecret      string
	APIKey             string
	ExportHandler      *handler.ExportHandler
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
	UsageHandler       *handler.UsageHandler
	AdminStatsHandler  *handler.AdminStatsHandler
	DashboardHandler   *handler.DashboardHandler
	InviteHandler      *handler.InviteHandler
	WaitlistHandler    *handler.WaitlistHandler
	SessionAuth        *auth.SessionAuth
	LLMKeyHandler        *handler.LLMKeyHandler
	CronHandler          *handler.CronHandler
	AdminWebhookHandler  *handler.AdminWebhookHandler
	AutoArchiveHandler   *handler.AutoArchiveHandler
	LabelHandler         *handler.LabelHandler
	SavedSearchHandler       *handler.SavedSearchHandler
	WeeklySummaryHandler     *handler.WeeklySummaryHandler
	EventsHandler            *handler.EventsHandler
	UserRateLimiter      *UserRateLimiter
	AuditRepo            *database.AuditRepo
}

func NewRouter(deps RouterDeps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth(deps.Pool))
	mux.HandleFunc("GET /docs", handleDocs)
	mux.HandleFunc("GET /api/openapi.yaml", handleOpenAPISpec)

	// Webhook (shared secret auth)
	webhookAuth := auth.WebhookAuth(deps.WebhookSecret, http.HandlerFunc(deps.WebhookHandler.IngestEmail))
	mux.Handle("POST /webhook/email", webhookAuth)

	// Cron endpoint (same shared secret auth, called by Cloud Scheduler)
	cronAuth := auth.WebhookAuth(deps.WebhookSecret, http.HandlerFunc(deps.CronHandler.TriggerDigests))
	mux.Handle("POST /cron/digests", cronAuth)

	// API routes (session auth) — public endpoints registered here but skipped by auth
	api := http.NewServeMux()

	// Public endpoints (no auth — skipped by publicPathAuth wrapper below)
	api.HandleFunc("GET /api/public/digests/{token}", deps.DigestHandler.GetPublic)
	api.HandleFunc("POST /api/public/unsubscribe", deps.PreferenceHandler.Unsubscribe)
	api.HandleFunc("GET /api/public/waitlist-status", deps.WaitlistHandler.Status)
	api.HandleFunc("GET /api/dashboard", deps.DashboardHandler.Get)
	api.HandleFunc("GET /api/emails", deps.EmailHandler.List)
	api.HandleFunc("PATCH /api/emails/batch", deps.EmailHandler.BatchUpdateState)
	api.HandleFunc("DELETE /api/emails/batch", deps.EmailHandler.BatchDelete)
	api.HandleFunc("GET /api/emails/{id}", deps.EmailHandler.Get)
	api.HandleFunc("PATCH /api/emails/{id}", deps.EmailHandler.UpdateState)
	api.HandleFunc("DELETE /api/emails/{id}", deps.EmailHandler.Delete)
	api.HandleFunc("POST /api/emails/{id}/reprocess", deps.EmailHandler.Reprocess)
	api.HandleFunc("POST /api/emails/{id}/unsubscribe", deps.EmailHandler.Unsubscribe)
	api.HandleFunc("GET /api/emails/{id}/thread", deps.EmailHandler.GetThread)
	api.HandleFunc("GET /api/subscriptions", deps.EmailHandler.ListSubscriptions)
	api.HandleFunc("POST /api/subscriptions/unsubscribe", deps.EmailHandler.BatchUnsubscribe)
	api.HandleFunc("GET /api/digests", deps.DigestHandler.List)
	api.HandleFunc("GET /api/digests/{id}", deps.DigestHandler.Get)
	api.HandleFunc("GET /api/weekly-summaries", deps.WeeklySummaryHandler.List)
	api.HandleFunc("GET /api/weekly-summaries/{id}", deps.WeeklySummaryHandler.Get)
	api.HandleFunc("POST /api/weekly-summaries/generate", deps.WeeklySummaryHandler.Generate)
	api.HandleFunc("DELETE /api/digests/{id}", deps.DigestHandler.Delete)
	api.HandleFunc("POST /api/digests/{id}/share", deps.DigestHandler.Share)
	api.HandleFunc("DELETE /api/digests/{id}/share", deps.DigestHandler.Unshare)
	api.HandleFunc("GET /api/accounts", deps.AccountHandler.List)
	api.HandleFunc("POST /api/accounts", deps.AccountHandler.Create)
	api.HandleFunc("DELETE /api/accounts/{id}", deps.AccountHandler.Delete)
	api.HandleFunc("GET /api/accounts/check-username", deps.AccountHandler.CheckUsername)
	api.HandleFunc("GET /api/preferences", deps.PreferenceHandler.Get)
	api.HandleFunc("PATCH /api/preferences", deps.PreferenceHandler.Update)
	api.HandleFunc("POST /api/preferences/test-webhook", deps.PreferenceHandler.TestWebhook)
	api.HandleFunc("GET /api/senders", deps.SenderHandler.List)
	api.HandleFunc("GET /api/senders/detail", deps.SenderHandler.GetDetail)
	api.HandleFunc("GET /api/senders/suggestions", deps.SenderHandler.Suggestions)
	api.HandleFunc("GET /api/senders/history", deps.SenderHandler.History)
	api.HandleFunc("PATCH /api/senders", deps.SenderHandler.Update)
	api.HandleFunc("GET /api/stats", deps.StatsHandler.Get)
	api.HandleFunc("GET /api/stats/history", deps.StatsHistoryHandler.Get)
	api.HandleFunc("POST /api/digests/preview", deps.DigestHandler.Preview)
	api.HandleFunc("POST /api/digests/generate", deps.DigestHandler.Generate)
	api.HandleFunc("GET /api/topics", deps.TopicHandler.List)
	api.HandleFunc("GET /api/usage", deps.UsageHandler.Get)
	api.HandleFunc("GET /api/llm-keys", deps.LLMKeyHandler.List)
	api.HandleFunc("PUT /api/llm-keys/{provider}", deps.LLMKeyHandler.Put)
	api.HandleFunc("PATCH /api/llm-keys/{provider}", deps.LLMKeyHandler.Patch)
	api.HandleFunc("DELETE /api/llm-keys/{provider}", deps.LLMKeyHandler.Delete)
	api.HandleFunc("GET /api/export", deps.ExportHandler.Export)
	api.HandleFunc("GET /api/labels", deps.LabelHandler.List)
	api.HandleFunc("POST /api/labels", deps.LabelHandler.Create)
	api.HandleFunc("PATCH /api/labels/{id}", deps.LabelHandler.Update)
	api.HandleFunc("DELETE /api/labels/{id}", deps.LabelHandler.Delete)
	api.HandleFunc("GET /api/emails/{id}/labels", deps.LabelHandler.ListByEmail)
	api.HandleFunc("PUT /api/emails/{id}/labels/{labelId}", deps.LabelHandler.AddToEmail)
	api.HandleFunc("DELETE /api/emails/{id}/labels/{labelId}", deps.LabelHandler.RemoveFromEmail)
	api.HandleFunc("POST /api/emails/labels/batch-add", deps.LabelHandler.BatchAdd)
	api.HandleFunc("POST /api/emails/labels/batch-remove", deps.LabelHandler.BatchRemove)
	api.HandleFunc("GET /api/saved-searches", deps.SavedSearchHandler.List)
	api.HandleFunc("POST /api/saved-searches", deps.SavedSearchHandler.Create)
	api.HandleFunc("DELETE /api/saved-searches/{id}", deps.SavedSearchHandler.Delete)
	api.HandleFunc("GET /api/auto-archive-rules", deps.AutoArchiveHandler.List)
	api.HandleFunc("PUT /api/auto-archive-rules", deps.AutoArchiveHandler.Upsert)
	api.HandleFunc("DELETE /api/auto-archive-rules/{id}", deps.AutoArchiveHandler.Delete)
	api.HandleFunc("POST /api/emails/{id}/feedback", deps.FeedbackHandler.Upsert)
	api.HandleFunc("GET /api/emails/{id}/feedback", deps.FeedbackHandler.Get)
	api.HandleFunc("DELETE /api/emails/{id}/feedback", deps.FeedbackHandler.Delete)
	api.HandleFunc("GET /api/access", deps.InviteHandler.CheckAccess)
	api.HandleFunc("POST /api/waitlist", deps.WaitlistHandler.Submit)
	api.HandleFunc("POST /api/digests/{id}/feedback", deps.DigestHandler.UpsertFeedback)
	api.HandleFunc("GET /api/digests/{id}/feedback", deps.DigestHandler.GetFeedback)
	api.HandleFunc("DELETE /api/digests/{id}/feedback", deps.DigestHandler.DeleteFeedback)

	// SSE events stream
	if deps.EventsHandler != nil {
		api.HandleFunc("GET /api/events", deps.EventsHandler.Stream)
	}

	// Admin routes
	admin := http.NewServeMux()
	admin.HandleFunc("POST /api/admin/digests/{id}/send", deps.DigestHandler.SendEmail)
	admin.HandleFunc("POST /api/admin/digests/generate", deps.DigestHandler.Generate)
	admin.HandleFunc("GET /api/admin/stats", deps.AdminStatsHandler.Get)
	admin.HandleFunc("GET /api/admin/users", deps.AdminStatsHandler.ListUsers)
	admin.HandleFunc("POST /api/admin/invites", deps.InviteHandler.Create)
	admin.HandleFunc("GET /api/admin/invites", deps.InviteHandler.List)
	admin.HandleFunc("GET /api/admin/waitlist", deps.WaitlistHandler.List)
	admin.HandleFunc("POST /api/admin/waitlist/{id}/approve", deps.WaitlistHandler.Approve)
	admin.HandleFunc("PATCH /api/admin/users/{id}/token-limit", deps.AdminStatsHandler.SetUserTokenLimit)
	admin.HandleFunc("PATCH /api/admin/settings/token-limit", deps.AdminStatsHandler.SetDefaultTokenLimit)
	admin.HandleFunc("GET /api/admin/settings/waitlist", deps.AdminStatsHandler.GetWaitlistEnabled)
	admin.HandleFunc("PATCH /api/admin/settings/waitlist", deps.AdminStatsHandler.SetWaitlistEnabled)
	if deps.AdminWebhookHandler != nil {
		admin.HandleFunc("GET /api/admin/emails/failed", deps.AdminWebhookHandler.ListFailed)
		admin.HandleFunc("POST /api/admin/emails/{id}/retry", deps.AdminWebhookHandler.RetryOne)
		admin.HandleFunc("POST /api/admin/emails/batch-retry", deps.AdminWebhookHandler.BatchRetry)
		admin.HandleFunc("POST /api/admin/webhooks/{id}/replay", deps.AdminWebhookHandler.Replay)
		admin.HandleFunc("GET /api/admin/pipeline", deps.AdminWebhookHandler.PipelineHealth)
	}
	admin.HandleFunc("GET /api/admin/audit-log", deps.AdminStatsHandler.GetAuditLog)

	var adminHandler http.Handler = admin
	if deps.AuditRepo != nil {
		adminHandler = AuditMiddleware(deps.AuditRepo)(admin)
	}
	api.Handle("/api/admin/", deps.SessionAuth.AdminOnly(adminHandler))

	var authedHandler http.Handler = api
	if deps.UserRateLimiter != nil {
		// Chain: SessionAuth (sets userID) → UserRateLimiter (checks userID) → handler
		authedHandler = deps.UserRateLimiter.Middleware(api)
	}
	authedAPI := publicPathSkip(auth.APIKeyAuth(deps.APIKey, deps.SessionAuth.Middleware(authedHandler)), api)
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

		// CSP: restrict sources. The /docs page loads Scalar from CDN so it needs
		// a looser policy; all other routes get a strict default.
		if r.URL.Path == "/docs" {
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; script-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'; "+
					"style-src 'self' https://cdn.jsdelivr.net 'unsafe-inline'; "+
					"img-src 'self' data: https:; font-src 'self' https://cdn.jsdelivr.net; "+
					"connect-src 'self'")
		} else {
			w.Header().Set("Content-Security-Policy",
				"default-src 'none'; frame-ancestors 'none'")
		}

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
				slog.Warn("health check failed", "error", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				// Don't expose internal details to unauthenticated callers
				json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
