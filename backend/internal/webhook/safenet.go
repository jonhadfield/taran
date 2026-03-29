package webhook

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ValidateExternalURL performs a fast pre-flight check that a URL is safe to
// make outbound requests to. Blocks private IPs, loopback, link-local, and
// GCP/AWS metadata endpoints. Note: this is a fast-fail optimisation — the
// authoritative check is in SafeHTTPClient's DialContext which validates IPs
// at connection time, preventing DNS rebinding attacks.
func ValidateExternalURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %q", scheme)
	}

	host := parsed.Hostname()

	// Block GCP metadata service hostnames
	if strings.EqualFold(host, "metadata.google.internal") ||
		strings.EqualFold(host, "metadata") {
		return fmt.Errorf("metadata service access blocked")
	}

	// Block AWS/GCP metadata IP
	if host == "169.254.169.254" {
		return fmt.Errorf("metadata service access blocked")
	}

	// Resolve hostname and check all IPs
	ips, err := net.LookupIP(host)
	if err != nil {
		ip := net.ParseIP(host)
		if ip != nil {
			if isPrivateIP(ip) {
				return fmt.Errorf("private/internal IP address blocked")
			}
			return nil
		}
		return fmt.Errorf("DNS resolution failed: %w", err)
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("resolved to private/internal IP address %s", ip)
		}
	}

	return nil
}

// SafeHTTPClient returns an http.Client that validates resolved IPs at
// connection time, preventing DNS rebinding attacks (TOCTOU). This is the
// authoritative SSRF gate — use this for all outbound requests to
// user-controlled URLs.
func SafeHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, fmt.Errorf("invalid address: %w", err)
				}

				// Block metadata hostnames at dial time too
				if strings.EqualFold(host, "metadata.google.internal") ||
					strings.EqualFold(host, "metadata") {
					return nil, fmt.Errorf("metadata service access blocked")
				}

				ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
				if err != nil {
					return nil, fmt.Errorf("DNS resolution failed: %w", err)
				}

				// Validate ALL resolved IPs before connecting
				for _, ip := range ips {
					if isPrivateIP(ip.IP) {
						return nil, fmt.Errorf("blocked private IP %s for host %s", ip.IP, host)
					}
				}

				if len(ips) == 0 {
					return nil, fmt.Errorf("no IPs resolved for host %s", host)
				}

				// Connect to the first valid IP directly (bypasses second DNS lookup)
				return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(
					ctx, network, net.JoinHostPort(ips[0].IP.String(), port),
				)
			},
		},
		// Block redirects to internal IPs and metadata endpoints
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			host := req.URL.Hostname()
			if strings.EqualFold(host, "metadata.google.internal") ||
				strings.EqualFold(host, "metadata") ||
				host == "169.254.169.254" {
				return fmt.Errorf("redirect to metadata service blocked")
			}
			// Also check resolved IPs for the redirect target
			if ip := net.ParseIP(host); ip != nil && isPrivateIP(ip) {
				return fmt.Errorf("redirect to private IP %s blocked", ip)
			}
			return nil
		},
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if ip.IsPrivate() {
		return true
	}
	if ip.IsUnspecified() {
		return true
	}
	return false
}
