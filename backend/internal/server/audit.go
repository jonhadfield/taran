package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
)

// AuditMiddleware logs admin write operations (POST, PATCH, PUT, DELETE) to the audit log.
// Read operations (GET) are not logged to avoid noise.
func AuditMiddleware(auditRepo *database.AuditRepo, resolver *ClientIPResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit mutating operations
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Capture response status
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			// Only log successful operations (2xx)
			if rec.status >= 200 && rec.status < 300 {
				userID := auth.UserIDFromContext(r.Context())
				userEmail := auth.UserEmailFromContext(r.Context())
				action := r.Method + " " + r.URL.Path
				target := extractTarget(r.URL.Path)
				ip := resolver.ClientIP(r)

				if err := auditRepo.Log(r.Context(), userID, userEmail, action, target, "", ip); err != nil {
					slog.Warn("failed to write audit log", "error", err, "action", action)
				}
			}
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush and Unwrap keep this wrapper transparent to streaming handlers and
// http.ResponseController. See statusWriter in middleware.go.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// extractTarget pulls the resource ID from admin paths like /api/admin/users/{id}/token-limit
func extractTarget(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/api/admin/"), "/")
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1] // e.g. "users/abc-123" or "digests/xyz-456"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return ""
}
