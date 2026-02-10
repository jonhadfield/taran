package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestWebhookAuth_ValidSecret(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := auth.WebhookAuth("my-secret", inner)
	req := httptest.NewRequest("POST", "/webhook/email", nil)
	req.Header.Set("X-Webhook-Secret", "my-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Error("inner handler was not called")
	}
}

func TestWebhookAuth_InvalidSecret(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := auth.WebhookAuth("my-secret", inner)
	req := httptest.NewRequest("POST", "/webhook/email", nil)
	req.Header.Set("X-Webhook-Secret", "wrong-secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Error("inner handler should not be called")
	}
}

func TestWebhookAuth_MissingSecret(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := auth.WebhookAuth("my-secret", inner)
	req := httptest.NewRequest("POST", "/webhook/email", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Error("inner handler should not be called")
	}
}

func TestSessionAuth_ValidBearerToken(t *testing.T) {
	sessions := &testutil.MockSessionRepo{
		GetByTokenFn: func(_ context.Context, token string) (*domain.Session, error) {
			return &domain.Session{
				ID:        "sess-1",
				UserID:    "user-123",
				Token:     token,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	sa := &auth.SessionAuth{Sessions: sessions}

	var gotUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = auth.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	sa.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != "user-123" {
		t.Errorf("userID = %q, want %q", gotUserID, "user-123")
	}
}

func TestSessionAuth_ValidCookie(t *testing.T) {
	sessions := &testutil.MockSessionRepo{
		GetByTokenFn: func(_ context.Context, token string) (*domain.Session, error) {
			return &domain.Session{
				ID:        "sess-1",
				UserID:    "user-456",
				Token:     token,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	sa := &auth.SessionAuth{Sessions: sessions}

	var gotUserID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = auth.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "cookie-token"})
	rec := httptest.NewRecorder()

	sa.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != "user-456" {
		t.Errorf("userID = %q, want %q", gotUserID, "user-456")
	}
}

func TestSessionAuth_ExpiredSession(t *testing.T) {
	sessions := &testutil.MockSessionRepo{
		GetByTokenFn: func(_ context.Context, token string) (*domain.Session, error) {
			return &domain.Session{
				ID:        "sess-1",
				UserID:    "user-123",
				Token:     token,
				ExpiresAt: time.Now().Add(-time.Hour),
			}, nil
		},
	}
	sa := &auth.SessionAuth{Sessions: sessions}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rec := httptest.NewRecorder()

	sa.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "session expired" {
		t.Errorf("error = %q, want %q", body["error"], "session expired")
	}
}

func TestSessionAuth_InvalidToken(t *testing.T) {
	sessions := &testutil.MockSessionRepo{
		GetByTokenFn: func(_ context.Context, token string) (*domain.Session, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	sa := &auth.SessionAuth{Sessions: sessions}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	sa.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSessionAuth_MissingToken(t *testing.T) {
	sessions := &testutil.MockSessionRepo{}
	sa := &auth.SessionAuth{Sessions: sessions}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	sa.Middleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "missing session token" {
		t.Errorf("error = %q, want %q", body["error"], "missing session token")
	}
}

func TestUserIDFromContext_Present(t *testing.T) {
	ctx := context.WithValue(context.Background(), auth.UserIDKey, "user-abc")
	got := auth.UserIDFromContext(ctx)
	if got != "user-abc" {
		t.Errorf("got %q, want %q", got, "user-abc")
	}
}

func TestUserIDFromContext_Missing(t *testing.T) {
	got := auth.UserIDFromContext(context.Background())
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
