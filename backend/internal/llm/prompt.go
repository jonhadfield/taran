package llm

import "fmt"

const extractionSystemPrompt = `You are an email analysis assistant. Analyze the following email and extract structured information.

Respond with a JSON object containing exactly these fields:
- summary: 1-3 sentence plain-language summary
- key_points: array of main takeaways (max 5)
- topics: array of category tags (max 5)
- links: array of {"url": "...", "title": "..."} for important URLs mentioned
- action_items: array of tasks, deadlines, or calls-to-action mentioned
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

func buildDigestUserPrompt(summaries []string, periodType string) string {
	prompt := fmt.Sprintf("Period: %s digest\n\nEmail summaries:\n\n", periodType)
	for i, s := range summaries {
		prompt += fmt.Sprintf("--- Email %d ---\n%s\n\n", i+1, s)
	}
	return prompt
}
