package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_AllowsNormalTraffic(t *testing.T) {
	rl := NewRateLimiter(10, 10, testResolver(t))
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_BlocksExcessTraffic(t *testing.T) {
	rl := NewRateLimiter(1, 2, testResolver(t)) // 1 req/s, burst of 2
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_SeparatesByIP(t *testing.T) {
	rl := NewRateLimiter(1, 1, testResolver(t))
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP uses its token
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "1.1.1.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("ip1 first request: status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Second IP should still be allowed
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "2.2.2.2:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("ip2 first request: status = %d, want %d", rec2.Code, http.StatusOK)
	}

	// First IP should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "1.1.1.1:1234"
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("ip1 second request: status = %d, want %d", rec3.Code, http.StatusTooManyRequests)
	}
}

func testResolver(t *testing.T) *ClientIPResolver {
	t.Helper()
	r, err := NewClientIPResolver(nil)
	if err != nil {
		t.Fatalf("NewClientIPResolver: %v", err)
	}
	return r
}

func TestClientIP_CFConnectingIPTrustedFromCloudflare(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("CF-Connecting-IP", "10.0.0.1")
	// 104.16.0.1 is inside Cloudflare's published 104.16.0.0/13 range.
	req.RemoteAddr = "104.16.0.1:1234"

	if got := testResolver(t).ClientIP(req); got != "10.0.0.1" {
		t.Errorf("ClientIP = %q, want %q", got, "10.0.0.1")
	}
}

// Regression: CF-Connecting-IP used to be trusted unconditionally, so any
// client reaching the origin directly could forge it — getting a fresh rate
// limit bucket per request and writing a chosen address into the audit log.
func TestClientIP_CFConnectingIPIgnoredFromUntrustedPeer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("CF-Connecting-IP", "10.0.0.1")
	req.RemoteAddr = "203.0.113.7:1234"

	if got := testResolver(t).ClientIP(req); got != "203.0.113.7" {
		t.Errorf("ClientIP = %q, want the peer address %q — a spoofed header must not win",
			got, "203.0.113.7")
	}
}

func TestClientIP_TrustNoProxy(t *testing.T) {
	resolver, err := NewClientIPResolver([]string{"none"})
	if err != nil {
		t.Fatalf("NewClientIPResolver: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("CF-Connecting-IP", "10.0.0.1")
	req.RemoteAddr = "104.16.0.1:1234"

	if got := resolver.ClientIP(req); got != "104.16.0.1" {
		t.Errorf("ClientIP = %q, want %q", got, "104.16.0.1")
	}
}

func TestClientIP_CustomTrustedProxy(t *testing.T) {
	resolver, err := NewClientIPResolver([]string{"192.168.1.0/24"})
	if err != nil {
		t.Fatalf("NewClientIPResolver: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("CF-Connecting-IP", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:1234"

	if got := resolver.ClientIP(req); got != "10.0.0.1" {
		t.Errorf("ClientIP = %q, want %q", got, "10.0.0.1")
	}
}

func TestNewClientIPResolver_RejectsInvalidCIDR(t *testing.T) {
	if _, err := NewClientIPResolver([]string{"not-a-cidr"}); err == nil {
		t.Error("expected an error for an invalid CIDR")
	}
}

func TestClientIP_XRealIPIgnored(t *testing.T) {
	// X-Real-Ip is intentionally ignored (spoofable) — should fall through to RemoteAddr
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-Ip", "10.0.0.2")
	req.RemoteAddr = "192.168.1.1:1234"

	if got := testResolver(t).ClientIP(req); got != "192.168.1.1" {
		t.Errorf("ClientIP = %q, want %q (X-Real-Ip should be ignored)", got, "192.168.1.1")
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	if got := testResolver(t).ClientIP(req); got != "192.168.1.1" {
		t.Errorf("ClientIP = %q, want %q", got, "192.168.1.1")
	}
}
