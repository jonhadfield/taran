package llm

import "fmt"

const extractionSystemPrompt = `You are an email analysis assistant. Analyze the following email and extract structured information.

The email content may be in markdown format converted from HTML. Use the structure (headings, links, lists, emphasis) to understand the content hierarchy and importance.

IMPORTANT: Ignore advertisements, sponsored content, affiliate promotions, tracking links, unsubscribe footers, and boilerplate legal text. Focus only on the primary editorial content of the email.

Respond with a JSON object containing exactly these fields:
- summary: 1-3 sentence plain-language summary of the primary content (excluding ads)
- key_points: array of main takeaways (max 5)
- topics: array of category tags (max 5)
- links: array of {"url": "...", "title": "..."} for editorially relevant URLs (exclude ad/tracking links)
- action_items: array of tasks, deadlines, or calls-to-action from the primary content
- sentiment: one of "informational", "urgent", "promotional", "personal", "transactional"
- source_category: one of "newsletter", "personal", "transactional", "marketing", "notification", "other"

Respond ONLY with valid JSON. No markdown fences, no explanation.`

func buildExtractionUserPrompt(subject, content, fromAddress string) string {
	return fmt.Sprintf("Subject: %s\nFrom: %s\n\n%s", subject, fromAddress, content)
}

const digestSystemPrompt = `You are a digest summarization assistant. Given multiple email extraction summaries, create a unified digest.

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
- Auto-replies and out-of-office messages
- Delivery status notifications (bounces, failures)
- Unsubscribe confirmations
- Password reset emails
- Two-factor authentication codes
- Pure spam or phishing attempts
- Calendar invitations with no substantive content
- Read receipts

EXTRACT these types of emails (extract: true):
- Newsletters with editorial content
- Curated digests and roundups
- Personal emails with substantive content
- Industry updates and reports
- Blog post notifications with content
- Product announcements with detail

Respond ONLY with valid JSON: {"extract": true/false, "reason": "brief reason"}
No markdown fences, no explanation outside the JSON.`

func buildTriageUserPrompt(subject, fromAddress, contentPreview string) string {
	return fmt.Sprintf("Subject: %s\nFrom: %s\n\nContent preview:\n%s", subject, fromAddress, contentPreview)
}

func buildDigestUserPrompt(summaries []string, periodType string) string {
	prompt := fmt.Sprintf("Period: %s digest\n\nEmail summaries:\n\n", periodType)
	for i, s := range summaries {
		prompt += fmt.Sprintf("--- Email %d ---\n%s\n\n", i+1, s)
	}
	return prompt
}
