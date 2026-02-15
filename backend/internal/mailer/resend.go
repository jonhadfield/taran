package mailer

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/resend/resend-go/v2"
)

type ResendMailer struct {
	client      *resend.Client
	fromAddress string
}

func NewResendMailer(apiKey, fromAddress string) *ResendMailer {
	return &ResendMailer{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
	}
}

func (m *ResendMailer) SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest) error {
	htmlBody := buildDigestHTML(digest)

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: digest.Title,
		Html:    htmlBody,
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send digest email: %w", err)
	}
	return nil
}

func buildDigestHTML(digest *domain.Digest) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#1a1a1a;">`)

	// Title
	b.WriteString(`<h1 style="font-size:24px;margin-bottom:8px;">`)
	b.WriteString(html.EscapeString(digest.Title))
	b.WriteString(`</h1>`)

	// Period
	b.WriteString(`<p style="color:#666;font-size:14px;margin-top:0;">`)
	b.WriteString(digest.PeriodStart.Format("Jan 2"))
	b.WriteString(` – `)
	b.WriteString(digest.PeriodEnd.Format("Jan 2, 2006"))
	b.WriteString(fmt.Sprintf(` · %d emails`, digest.EmailCount))
	b.WriteString(`</p>`)

	// Summary
	b.WriteString(`<p style="font-size:16px;line-height:1.5;">`)
	b.WriteString(html.EscapeString(digest.Summary))
	b.WriteString(`</p>`)

	// Highlights
	if len(digest.Highlights) > 0 {
		b.WriteString(`<h2 style="font-size:18px;margin-top:24px;">Highlights</h2>`)
		b.WriteString(`<ul style="padding-left:20px;line-height:1.6;">`)
		for _, h := range digest.Highlights {
			b.WriteString(`<li>`)
			b.WriteString(html.EscapeString(h))
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul>`)
	}

	// Topics
	if len(digest.TopTopics) > 0 {
		b.WriteString(`<div style="margin-top:24px;">`)
		for _, topic := range digest.TopTopics {
			b.WriteString(`<span style="display:inline-block;background:#f0f0f0;border-radius:12px;padding:4px 12px;margin:2px 4px;font-size:13px;">`)
			b.WriteString(html.EscapeString(topic))
			b.WriteString(`</span>`)
		}
		b.WriteString(`</div>`)
	}

	// Email summaries
	if len(digest.EmailSummaries) > 0 {
		b.WriteString(`<h2 style="font-size:18px;margin-top:32px;margin-bottom:16px;">Email Summaries</h2>`)
		for _, es := range digest.EmailSummaries {
			b.WriteString(`<div style="margin-bottom:16px;padding:12px;border:1px solid #eee;border-radius:8px;">`)

			// Subject + sender
			b.WriteString(`<div style="font-weight:600;font-size:14px;margin-bottom:4px;">`)
			b.WriteString(html.EscapeString(es.Subject))
			b.WriteString(`</div>`)
			if es.SenderName != "" {
				b.WriteString(`<div style="font-size:12px;color:#666;margin-bottom:8px;">`)
				b.WriteString(html.EscapeString(es.SenderName))
				b.WriteString(`</div>`)
			}

			// Summary
			b.WriteString(`<div style="font-size:13px;line-height:1.4;color:#333;">`)
			b.WriteString(html.EscapeString(es.Summary))
			b.WriteString(`</div>`)

			// Link to dashboard
			b.WriteString(`<a href="https://mailbrief.io/inbox/`)
			b.WriteString(html.EscapeString(es.EmailID))
			b.WriteString(`" style="display:inline-block;margin-top:8px;font-size:12px;color:#0066cc;text-decoration:none;">View in dashboard →</a>`)

			b.WriteString(`</div>`)
		}
	}

	// Footer
	b.WriteString(`<hr style="border:none;border-top:1px solid #eee;margin-top:32px;">`)
	b.WriteString(`<p style="color:#999;font-size:12px;">Sent by <a href="https://mailbrief.io" style="color:#999;">MailBrief</a>. `)
	b.WriteString(`You can disable email delivery in your <a href="https://mailbrief.io/settings" style="color:#999;">settings</a>.</p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}
