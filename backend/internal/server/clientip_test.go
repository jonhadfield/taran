package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/server"
)

const testAPIKey = "test-api-key"

func request(t *testing.T, peer string, headers map[string]string) *http.Request {
	t.Helper()
	r := httptest.NewRequest("GET", "/api/emails", nil)
	r.RemoteAddr = peer
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func resolver(t *testing.T, cidrs []string, apiKey string) *server.ClientIPResolver {
	t.Helper()
	r, err := server.NewClientIPResolver(cidrs)
	if err != nil {
		t.Fatal(err)
	}
	if apiKey != "" {
		r.TrustAPIKeyClientIP(apiKey)
	}
	return r
}

// The frontend proxies every browser call, so without this each user would be
// rate-limited against the frontend's address rather than their own.
func TestClientIP_HonoursHeaderFromAPIKeyHolder(t *testing.T) {
	r := resolver(t, []string{"none"}, testAPIKey)
	req := request(t, "10.1.2.3:443", map[string]string{
		"X-API-Key":           testAPIKey,
		server.ClientIPHeader: "203.0.113.9",
	})
	if got := r.ClientIP(req); got != "203.0.113.9" {
		t.Errorf("ClientIP = %q, want the declared end-user address", got)
	}
}

func TestClientIP_IgnoresHeaderWithoutTheAPIKey(t *testing.T) {
	r := resolver(t, []string{"none"}, testAPIKey)
	tests := map[string]map[string]string{
		"no key":      {server.ClientIPHeader: "203.0.113.9"},
		"wrong key":   {"X-API-Key": "guessed", server.ClientIPHeader: "203.0.113.9"},
		"bad address": {"X-API-Key": testAPIKey, server.ClientIPHeader: "not-an-ip"},
	}
	for name, headers := range tests {
		if got := r.ClientIP(request(t, "198.51.100.7:443", headers)); got != "198.51.100.7" {
			t.Errorf("%s: ClientIP = %q, want the peer address", name, got)
		}
	}
}

func TestClientIP_DisabledUntilTrustConfigured(t *testing.T) {
	r := resolver(t, []string{"none"}, "")
	req := request(t, "198.51.100.7:443", map[string]string{
		"X-API-Key":           testAPIKey,
		server.ClientIPHeader: "203.0.113.9",
	})
	if got := r.ClientIP(req); got != "198.51.100.7" {
		t.Errorf("ClientIP = %q, want the peer address", got)
	}
}

func TestClientIP_CloudflareHeaderStillHonouredFromTrustedRange(t *testing.T) {
	r := resolver(t, []string{"162.158.0.0/15"}, testAPIKey)

	fromCloudflare := request(t, "162.158.1.1:443", map[string]string{"CF-Connecting-IP": "203.0.113.9"})
	if got := r.ClientIP(fromCloudflare); got != "203.0.113.9" {
		t.Errorf("trusted proxy: ClientIP = %q, want 203.0.113.9", got)
	}

	spoofed := request(t, "198.51.100.7:443", map[string]string{"CF-Connecting-IP": "203.0.113.9"})
	if got := r.ClientIP(spoofed); got != "198.51.100.7" {
		t.Errorf("untrusted peer: ClientIP = %q, want the peer address", got)
	}
}
