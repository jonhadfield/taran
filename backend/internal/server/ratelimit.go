package server

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"golang.org/x/time/rate"
)

// keyedRateLimiter is the shared implementation for both IP-based and user-based
// rate limiting. It tracks per-key request rates with automatic cleanup of stale entries.
type keyedRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientEntry
	rate     rate.Limit
	burst    int
	cleanTTL time.Duration
	stop     chan struct{}
}

type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newKeyedRateLimiter(rps float64, burst int) *keyedRateLimiter {
	krl := &keyedRateLimiter{
		clients:  make(map[string]*clientEntry),
		rate:     rate.Limit(rps),
		burst:    burst,
		cleanTTL: 3 * time.Minute,
		stop:     make(chan struct{}),
	}
	go krl.cleanup()
	return krl
}

func (krl *keyedRateLimiter) Stop() {
	close(krl.stop)
}

func (krl *keyedRateLimiter) allow(key string) bool {
	krl.mu.Lock()
	defer krl.mu.Unlock()

	entry, ok := krl.clients[key]
	if !ok {
		entry = &clientEntry{
			limiter: rate.NewLimiter(krl.rate, krl.burst),
		}
		krl.clients[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (krl *keyedRateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			krl.mu.Lock()
			cutoff := time.Now().Add(-krl.cleanTTL)
			for key, entry := range krl.clients {
				if entry.lastSeen.Before(cutoff) {
					delete(krl.clients, key)
				}
			}
			krl.mu.Unlock()
		case <-krl.stop:
			return
		}
	}
}

// RateLimiter tracks per-IP request rates.
type RateLimiter struct {
	*keyedRateLimiter
}

// NewRateLimiter creates a rate limiter with the given requests per second and burst size.
// Stale entries are cleaned up after 3 minutes of inactivity.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{keyedRateLimiter: newKeyedRateLimiter(rps, burst)}
}

// Middleware returns an HTTP middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		if !rl.allow(ip) {
			slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SplitRateLimiter applies different rate limits based on the request path.
// Webhook/cron paths get higher limits since they're machine-to-machine.
type SplitRateLimiter struct {
	api     *RateLimiter
	webhook *RateLimiter
}

// NewSplitRateLimiter creates separate rate limiters for API and webhook traffic.
func NewSplitRateLimiter(apiRPS float64, apiBurst int, webhookRPS float64, webhookBurst int) *SplitRateLimiter {
	return &SplitRateLimiter{
		api:     NewRateLimiter(apiRPS, apiBurst),
		webhook: NewRateLimiter(webhookRPS, webhookBurst),
	}
}

func (sl *SplitRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/webhook/") || strings.HasPrefix(path, "/cron/") {
			sl.webhook.Middleware(next).ServeHTTP(w, r)
		} else {
			sl.api.Middleware(next).ServeHTTP(w, r)
		}
	})
}

// UserRateLimiter applies per-user rate limits using the authenticated user ID
// from the request context. Runs after auth middleware.
type UserRateLimiter struct {
	*keyedRateLimiter
}

// NewUserRateLimiter creates a per-user rate limiter.
func NewUserRateLimiter(rps float64, burst int) *UserRateLimiter {
	return &UserRateLimiter{keyedRateLimiter: newKeyedRateLimiter(rps, burst)}
}

// Middleware returns an HTTP middleware that rate-limits by authenticated user ID.
// If no user ID is in context (unauthenticated), the request passes through.
func (rl *UserRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserIDFromContext(r.Context())
		if userID == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !rl.allow(userID) {
			slog.Warn("per-user rate limit exceeded", "userID", userID, "path", r.URL.Path)
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// Prefer CF-Connecting-IP (set by Cloudflare, cannot be spoofed by client)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// X-Real-Ip omitted — spoofable when not behind a trusted proxy.
	// Fall through to RemoteAddr (TCP connection source).

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
