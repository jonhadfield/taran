package webhook

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateExternalURL checks that a URL is safe to make outbound requests to.
// Blocks private IPs, loopback, link-local, and GCP metadata endpoints.
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

	// Block AWS metadata IP
	if host == "169.254.169.254" {
		return fmt.Errorf("metadata service access blocked")
	}

	// Resolve hostname and check all IPs
	ips, err := net.LookupIP(host)
	if err != nil {
		// If DNS fails, also try parsing as a direct IP
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

func isPrivateIP(ip net.IP) bool {
	// Loopback (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		return true
	}

	// Link-local (169.254.0.0/16, fe80::/10)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Private ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7)
	if ip.IsPrivate() {
		return true
	}

	// Unspecified (0.0.0.0, ::)
	if ip.IsUnspecified() {
		return true
	}

	return false
}
