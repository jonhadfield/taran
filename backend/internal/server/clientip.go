package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// cloudflareRanges are Cloudflare's published edge network CIDRs. Requests
// arriving from these addresses may set CF-Connecting-IP; requests from
// anywhere else may not, because the header is trivially forged by any client
// that can reach the origin directly.
//
// Source: https://www.cloudflare.com/ips/
var cloudflareRanges = []string{
	// IPv4
	"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22",
	"141.101.64.0/18", "108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20",
	"197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
	"104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
	// IPv6
	"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32",
	"2405:8100::/32", "2a06:98c0::/29", "2c0f:f248::/32",
}

// ClientIPResolver determines the originating client IP for a request,
// honouring proxy headers only when the request actually came from a proxy we
// trust.
type ClientIPResolver struct {
	trusted []*net.IPNet
}

// NewClientIPResolver builds a resolver from a list of trusted proxy CIDRs.
// An empty list falls back to Cloudflare's published ranges. Pass a list
// containing "none" to trust no proxy at all, so only the TCP peer address is
// ever used.
func NewClientIPResolver(cidrs []string) (*ClientIPResolver, error) {
	if len(cidrs) == 1 && strings.EqualFold(strings.TrimSpace(cidrs[0]), "none") {
		return &ClientIPResolver{}, nil
	}
	if len(cidrs) == 0 {
		cidrs = cloudflareRanges
	}

	r := &ClientIPResolver{}
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		_, network, err := net.ParseCIDR(c)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", c, err)
		}
		r.trusted = append(r.trusted, network)
	}
	return r, nil
}

// ClientIP returns the best available client address for rate limiting and
// audit logging.
func (c *ClientIPResolver) ClientIP(r *http.Request) string {
	peer := peerIP(r)

	// Only a request that genuinely arrived from a trusted proxy may override
	// the peer address via CF-Connecting-IP. Otherwise an attacker sets the
	// header themselves and gets a fresh rate-limit bucket per request, and
	// writes an address of their choosing into the admin audit log.
	if c.isTrusted(peer) {
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			if ip := net.ParseIP(strings.TrimSpace(cfIP)); ip != nil {
				return ip.String()
			}
		}
	}

	return peer
}

func (c *ClientIPResolver) isTrusted(peer string) bool {
	ip := net.ParseIP(peer)
	if ip == nil {
		return false
	}
	for _, network := range c.trusted {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// peerIP is the address of the TCP connection itself, which cannot be forged.
func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
