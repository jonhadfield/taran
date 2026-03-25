package auth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"

	// tokenRotationAge is how long a token can be used before being rotated.
	tokenRotationAge = 1 * time.Hour
)

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

func UserEmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(UserEmailKey).(string)
	return email
}

// APIKeyAuth validates the X-API-Key header against the configured key.
func APIKeyAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("X-API-Key")
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
			writeAuthError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WebhookAuth validates the shared secret for webhook endpoints.
func WebhookAuth(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("X-Webhook-Secret")
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(secret)) != 1 {
			writeAuthError(w, http.StatusUnauthorized, "invalid webhook secret")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SessionAuth validates Better Auth session tokens for API endpoints.
type SessionAuth struct {
	Sessions    database.SessionRepository
	Invites     database.InviteRepository
	AdminEmails []string
}

func (a *SessionAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing session token")
			return
		}

		session, err := a.Sessions.GetByToken(r.Context(), token)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid session")
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			writeAuthError(w, http.StatusUnauthorized, "session expired")
			return
		}

		// Token rotation: if the token hasn't been updated in over an hour,
		// generate a new one and signal the frontend to swap the cookie.
		if time.Since(session.UpdatedAt) > tokenRotationAge {
			newToken := uuid.New().String()
			if err := a.Sessions.RotateToken(r.Context(), token, newToken); err != nil {
				slog.Warn("token rotation failed", "error", err)
			} else {
				w.Header().Set("X-Session-Token-Rotated", newToken)
			}
		}

		ctx := context.WithValue(r.Context(), UserIDKey, session.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, session.UserEmail)

		// Invite gate: enforce on backend (not just frontend)
		if a.Invites != nil && !a.isAllowed(r.Context(), session.UserEmail) {
			writeAuthError(w, http.StatusForbidden, "not invited")
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnly wraps a handler and rejects non-admin users.
func (a *SessionAuth) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := strings.ToLower(UserEmailFromContext(r.Context()))
		for _, admin := range a.AdminEmails {
			if email == admin {
				next.ServeHTTP(w, r)
				return
			}
		}
		writeAuthError(w, http.StatusForbidden, "admin access required")
	})
}

func (a *SessionAuth) isAllowed(ctx context.Context, email string) bool {
	email = strings.ToLower(email)
	for _, admin := range a.AdminEmails {
		if email == admin {
			return true
		}
	}
	if a.Invites != nil {
		invite, _ := a.Invites.GetByEmail(ctx, email)
		return invite != nil
	}
	return false
}

func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie("better-auth.session_token"); err == nil {
		return cookie.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
