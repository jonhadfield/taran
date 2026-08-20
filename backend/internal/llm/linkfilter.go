package llm

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/hadfielj/taran/backend/internal/domain"
)

// sourceURLPattern finds http(s) URLs in the original email content.
var sourceURLPattern = regexp.MustCompile(`https?://[^\s"'<>)\]]+`)

// filterLinksToSource drops any link the model returned whose host does not
// appear in the source email.
//
// The email body is untrusted and goes into the prompt verbatim, so a crafted
// message can try to talk the model into emitting a link of the attacker's
// choosing. Those links are rendered in the dashboard and summarised into
// digests, which would make the product a delivery vehicle for a URL that was
// never actually in the mail. Matching on host rather than exact URL keeps
// legitimate links that the model normalised or de-tracked.
func filterLinksToSource(links []domain.Link, content string) []domain.Link {
	if len(links) == 0 {
		return links
	}

	sourceHosts := make(map[string]bool)
	for _, raw := range sourceURLPattern.FindAllString(content, -1) {
		if h := normalisedHost(raw); h != "" {
			sourceHosts[h] = true
		}
	}

	kept := make([]domain.Link, 0, len(links))
	for _, link := range links {
		host := normalisedHost(link.URL)
		if host == "" || !sourceHosts[host] {
			continue
		}
		kept = append(kept, link)
	}
	return kept
}

// normalisedHost lowercases the host and strips a leading "www." so that
// cosmetic differences do not cause a legitimate link to be dropped.
func normalisedHost(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}
