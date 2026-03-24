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

// RateLimiter tracks per-IP request rates.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientEntry
	rate     rate.Limit
	burst    int
	cleanTTL time.Duration
}

type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter with the given requests per second and burst size.
// Stale entries are cleaned up after 3 minutes of inactivity.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*clientEntry),
		rate:     rate.Limit(rps),
		burst:    burst,
		cleanTTL: 3 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.clients[ip]
	if !ok {
		entry = &clientEntry{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.clients[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanTTL)
		for ip, entry := range rl.clients {
			if entry.lastSeen.Before(cutoff) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an HTTP middleware that rate-limits by client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
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
	mu       sync.Mutex
	clients  map[string]*clientEntry
	rate     rate.Limit
	burst    int
	cleanTTL time.Duration
}

// NewUserRateLimiter creates a per-user rate limiter.
func NewUserRateLimiter(rps float64, burst int) *UserRateLimiter {
	rl := &UserRateLimiter{
		clients:  make(map[string]*clientEntry),
		rate:     rate.Limit(rps),
		burst:    burst,
		cleanTTL: 3 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *UserRateLimiter) getLimiter(userID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.clients[userID]
	if !ok {
		entry = &clientEntry{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.clients[userID] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (rl *UserRateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanTTL)
		for id, entry := range rl.clients {
			if entry.lastSeen.Before(cutoff) {
				delete(rl.clients, id)
			}
		}
		rl.mu.Unlock()
	}
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

		limiter := rl.getLimiter(userID)
		if !limiter.Allow() {
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

	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
