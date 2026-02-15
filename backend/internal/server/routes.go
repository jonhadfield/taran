package server

import (
	"encoding/json"
	"net/http"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/handler"
)

type RouterDeps struct {
	WebhookSecret     string
	APIKey            string
	WebhookHandler    *handler.WebhookHandler
	EmailHandler      *handler.EmailHandler
	DigestHandler     *handler.DigestHandler
	AccountHandler    *handler.AccountHandler
	PreferenceHandler *handler.PreferenceHandler
	SessionAuth       *auth.SessionAuth
}

func NewRouter(deps RouterDeps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)

	// Public endpoints (no auth)
	mux.HandleFunc("GET /api/public/digests/{token}", deps.DigestHandler.GetPublic)

	// Webhook (shared secret auth)
	webhookAuth := auth.WebhookAuth(deps.WebhookSecret, http.HandlerFunc(deps.WebhookHandler.IngestEmail))
	mux.Handle("POST /webhook/email", webhookAuth)

	// API routes (session auth)
	api := http.NewServeMux()
	api.HandleFunc("GET /api/emails", deps.EmailHandler.List)
	api.HandleFunc("GET /api/emails/{id}", deps.EmailHandler.Get)
	api.HandleFunc("PATCH /api/emails/{id}", deps.EmailHandler.UpdateState)
	api.HandleFunc("GET /api/digests", deps.DigestHandler.List)
	api.HandleFunc("GET /api/digests/{id}", deps.DigestHandler.Get)
	api.HandleFunc("POST /api/digests/{id}/share", deps.DigestHandler.Share)
	api.HandleFunc("DELETE /api/digests/{id}/share", deps.DigestHandler.Unshare)
	api.HandleFunc("GET /api/accounts", deps.AccountHandler.List)
	api.HandleFunc("POST /api/accounts", deps.AccountHandler.Create)
	api.HandleFunc("DELETE /api/accounts/{id}", deps.AccountHandler.Delete)
	api.HandleFunc("GET /api/accounts/check-username", deps.AccountHandler.CheckUsername)
	api.HandleFunc("GET /api/preferences", deps.PreferenceHandler.Get)
	api.HandleFunc("PATCH /api/preferences", deps.PreferenceHandler.Update)

	// Admin routes
	admin := http.NewServeMux()
	admin.HandleFunc("POST /api/admin/digests/generate", deps.DigestHandler.Generate)
	api.Handle("/api/admin/", deps.SessionAuth.AdminOnly(admin))

	mux.Handle("/api/", auth.APIKeyAuth(deps.APIKey, deps.SessionAuth.Middleware(api)))

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
