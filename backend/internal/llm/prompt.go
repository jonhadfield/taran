package llm

import (
	"fmt"
	"strings"

	"github.com/hadfielj/taran/backend/internal/domain"
)

const extractionSystemPrompt = `You are an email analysis assistant. Analyze the following email and extract structured information.

The email content may be in markdown format converted from HTML. Use the structure (headings, links, lists, emphasis) to understand the content hierarchy and importance.

IMPORTANT: Ignore advertisements, sponsored content, affiliate promotions, tracking links, unsubscribe footers, and boilerplate legal text. Focus only on the primary editorial content of the email.

Respond with a JSON object containing exactly these fields:
- summary: 1-3 sentence plain-language summary of the primary content (excluding ads)
- key_points: array of main takeaways (max 5)
- topics: array of category tags (max 5)
- links: array of {"url": "...", "title": "..."} for editorially relevant URLs (max 5). IMPORTANT: Omit links that are tracking redirects, click-tracking wrappers, or URLs longer than 200 characters. If only tracking URLs are available, return an empty array.
- sentiment: one of "informational", "urgent", "promotional", "personal", "transactional"
- source_category: one of "newsletter", "personal", "transactional", "marketing", "notification", "other"

Respond ONLY with valid JSON. No markdown fences, no explanation.`

func buildExtractionUserPrompt(subject, content, fromAddress string) string {
	return fmt.Sprintf("Subject: %s\nFrom: %s\n\n%s", subject, fromAddress, content)
}

const digestSystemPrompt = `You are a digest summarization assistant. Given multiple email extraction summaries, create a unified digest.

When user topic preferences are provided, give more prominence to preferred topics in the highlights and summary. De-emphasize (but do not completely exclude) less preferred topics.

When user keyword preferences are provided, give strong prominence to content matching interest keywords. Completely omit content matching exclusion keywords.

Respond with a JSON object containing exactly these fields:
- title: a short descriptive title for this digest (e.g. "Daily Digest - Tech & Business")
- summary: 2-4 sentence overview of the most important themes across all emails
- highlights: array of the top 3-5 most noteworthy items across all emails
- top_topics: array of the most common topics (max 5)

Respond ONLY with valid JSON. No markdown fences, no explanation.`

const triageSystemPrompt = `You are an email triage assistant. Decide whether an email should be fully analyzed or skipped.

SKIP these types of emails (extract: false):
- Subscription confirmations ("confirm your subscription", "verify your email")
- Email verification / account activation emails
- Sign-up notifications ("welcome to...", "your account has been created")
- Auto-replies and out-of-office messages
- Delivery status notifications (bounces, failures)
- Unsubscribe confirmations
- Password reset emails
- Two-factor authentication codes
- Pure spam or phishing attempts
- Calendar invitations with no substantive content
- Read receipts
- Transactional emails (order confirmations, shipping updates, receipts)
- Marketing / promotional blasts with no editorial content

EXTRACT these types of emails (extract: true):
- Newsletters with editorial content
- Curated digests and roundups (e.g. Hacker News, TLDR, Morning Brew, link roundups)
- Personal emails with substantive content
- Industry updates and reports
- Blog post notifications with content
- Product announcements with detail
- Any email with multiple links and headlines (likely a newsletter or digest)

IMPORTANT: When in doubt, extract. It is much better to extract a low-value email than to skip a valuable newsletter. Only skip emails that are clearly automated/transactional with no editorial content.

Respond ONLY with valid JSON: {"extract": true/false, "reason": "brief reason"}
No markdown fences, no explanation outside the JSON.`

func buildTriageUserPrompt(subject, fromAddress, contentPreview string) string {
	return fmt.Sprintf("Subject: %s\nFrom: %s\n\nContent preview:\n%s", subject, fromAddress, contentPreview)
}

func buildDigestUserPrompt(extractions []domain.Extraction, periodType string, opts *DigestOptions) string {
	prompt := fmt.Sprintf("Period: %s digest\n\nEmail summaries:\n\n", periodType)
	for i, e := range extractions {
		prompt += fmt.Sprintf("--- Email %d ---\n%s\n", i+1, e.Summary)
		if len(e.Topics) > 0 {
			prompt += fmt.Sprintf("Topics: %s\n", strings.Join(e.Topics, ", "))
		}
		prompt += "\n"
	}

	if opts != nil && (len(opts.PreferredTopics) > 0 || len(opts.LessPreferredTopics) > 0) {
		prompt += "User topic preferences:\n"
		if len(opts.PreferredTopics) > 0 {
			prompt += fmt.Sprintf("- Preferred topics: %s\n", strings.Join(opts.PreferredTopics, ", "))
		}
		if len(opts.LessPreferredTopics) > 0 {
			prompt += fmt.Sprintf("- Less preferred topics: %s\n", strings.Join(opts.LessPreferredTopics, ", "))
		}
	}

	if opts != nil && (len(opts.InterestKeywords) > 0 || len(opts.ExclusionKeywords) > 0) {
		prompt += "User keyword preferences:\n"
		if len(opts.InterestKeywords) > 0 {
			prompt += fmt.Sprintf("- Interest keywords (boost these): %s\n", strings.Join(opts.InterestKeywords, ", "))
		}
		if len(opts.ExclusionKeywords) > 0 {
			prompt += fmt.Sprintf("- Exclusion keywords (omit these): %s\n", strings.Join(opts.ExclusionKeywords, ", "))
		}
	}

	if opts != nil && opts.Style == "concise" {
		prompt += "\nIMPORTANT: The user prefers a concise digest. Keep the summary to 1-2 sentences and limit highlights to 2-3 items maximum.\n"
	}

	return prompt
}
