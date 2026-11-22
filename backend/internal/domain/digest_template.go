package domain

// Digest section identifiers used in DigestSectionOrder preferences.
const (
	DigestSectionHighlights = "highlights"
	DigestSectionTopics     = "topics"
	DigestSectionEmails     = "emails"

	DigestDensityComfortable = "comfortable"
	DigestDensityCompact     = "compact"
)

// DefaultDigestSectionOrder is the legacy layout: highlights, then topics, then emails.
var DefaultDigestSectionOrder = []string{
	DigestSectionHighlights,
	DigestSectionTopics,
	DigestSectionEmails,
}

var validDigestSections = map[string]bool{
	DigestSectionHighlights: true,
	DigestSectionTopics:     true,
	DigestSectionEmails:     true,
}

// NormalizeDigestSectionOrder returns a de-duplicated order containing each
// known section exactly once. Unknown entries are dropped; missing ones are
// appended in the default order.
func NormalizeDigestSectionOrder(order []string) []string {
	seen := make(map[string]bool, len(DefaultDigestSectionOrder))
	out := make([]string, 0, len(DefaultDigestSectionOrder))
	for _, s := range order {
		if !validDigestSections[s] || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, s := range DefaultDigestSectionOrder {
		if !seen[s] {
			out = append(out, s)
		}
	}
	return out
}

// NormalizeDigestDensity returns a valid density, defaulting to comfortable.
func NormalizeDigestDensity(density string) string {
	if density == DigestDensityCompact {
		return DigestDensityCompact
	}
	return DigestDensityComfortable
}

// ApplyDigestRenderPrefs copies section order and density onto a digest for
// transient email/PDF rendering. Safe with a nil preference (uses defaults).
func ApplyDigestRenderPrefs(d *Digest, pref *UserPreference) {
	if d == nil {
		return
	}
	if pref == nil {
		d.SectionOrder = append([]string(nil), DefaultDigestSectionOrder...)
		d.Density = DigestDensityComfortable
		return
	}
	d.SectionOrder = NormalizeDigestSectionOrder(pref.DigestSectionOrder)
	d.Density = NormalizeDigestDensity(pref.DigestDensity)
}
