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

func (m *ResendMailer) SendDigest(ctx context.Context, toEmail, toName string, digest *domain.Digest, unsubscribeURL string) error {
	htmlBody := buildDigestHTML(digest, unsubscribeURL)

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: digest.Title,
		Html:    htmlBody,
	}

	if unsubscribeURL != "" {
		params.Headers = map[string]string{
			"List-Unsubscribe":      "<" + unsubscribeURL + ">",
			"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
		}
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send digest email: %w", err)
	}
	return nil
}

func (m *ResendMailer) SendInvite(ctx context.Context, toEmail, fromName string) error {
	htmlBody := buildInviteHTML(fromName)

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: "You're invited to MailBrief",
		Html:    htmlBody,
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send invite email: %w", err)
	}
	return nil
}

func buildInviteHTML(fromName string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#1a1a1a;">`)

	b.WriteString(`<h1 style="font-size:24px;margin-bottom:8px;">You're invited to MailBrief</h1>`)

	b.WriteString(`<p style="font-size:16px;line-height:1.5;">`)
	b.WriteString(html.EscapeString(fromName))
	b.WriteString(` has invited you to join MailBrief — AI-powered digests of your newsletters, delivered daily.</p>`)

	b.WriteString(`<p style="font-size:16px;line-height:1.5;">Sign in to get started:</p>`)

	b.WriteString(`<a href="https://mailbrief.io/login" style="display:inline-block;background:#0066cc;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-size:16px;font-weight:500;">Sign in to MailBrief</a>`)

	b.WriteString(`<hr style="border:none;border-top:1px solid #eee;margin-top:32px;">`)
	b.WriteString(`<p style="color:#999;font-size:12px;">Sent by <a href="https://mailbrief.io" style="color:#999;">MailBrief</a></p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}

var categoryColors = map[string]string{
	"newsletter":    "#1d4ed8",
	"personal":      "#059669",
	"transactional": "#9333ea",
	"marketing":     "#d97706",
	"notification":  "#6b7280",
	"other":         "#6b7280",
}

var categoryBgColors = map[string]string{
	"newsletter":    "#dbeafe",
	"personal":      "#d1fae5",
	"transactional": "#f3e8ff",
	"marketing":     "#fef3c7",
	"notification":  "#f3f4f6",
	"other":         "#f3f4f6",
}

func buildDigestHTML(digest *domain.Digest, unsubscribeURL string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:24px 16px;color:#1a1a1a;background:#ffffff;">`)

	// Header bar
	b.WriteString(`<table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:24px;"><tr>`)
	b.WriteString(`<td><span style="font-size:13px;font-weight:600;color:#6b7280;letter-spacing:0.05em;text-transform:uppercase;">MailBrief</span></td>`)
	b.WriteString(`<td align="right">`)
	if digest.ShareToken != nil {
		b.WriteString(`<a href="https://mailbrief.io/shared/`)
		b.WriteString(html.EscapeString(*digest.ShareToken))
		b.WriteString(`" style="font-size:12px;color:#0066cc;text-decoration:none;">View in browser &rarr;</a>`)
	}
	b.WriteString(`</td></tr></table>`)

	// Title
	b.WriteString(`<h1 style="font-size:22px;font-weight:700;margin:0 0 4px 0;line-height:1.3;">`)
	b.WriteString(html.EscapeString(digest.Title))
	b.WriteString(`</h1>`)

	// Period + count
	b.WriteString(`<p style="color:#6b7280;font-size:13px;margin:0 0 20px 0;">`)
	b.WriteString(digest.PeriodStart.Format("Jan 2"))
	b.WriteString(` – `)
	b.WriteString(digest.PeriodEnd.Format("Jan 2, 2006"))
	b.WriteString(fmt.Sprintf(` · %d emails`, digest.EmailCount))
	b.WriteString(`</p>`)

	// Summary
	b.WriteString(`<p style="font-size:15px;line-height:1.6;color:#374151;margin:0 0 24px 0;">`)
	b.WriteString(html.EscapeString(digest.Summary))
	b.WriteString(`</p>`)

	// Highlights
	if len(digest.Highlights) > 0 {
		b.WriteString(`<div style="background:#f8fafc;border-radius:8px;padding:16px;margin-bottom:24px;">`)
		b.WriteString(`<h2 style="font-size:14px;font-weight:600;color:#374151;margin:0 0 10px 0;text-transform:uppercase;letter-spacing:0.03em;">Key Highlights</h2>`)
		b.WriteString(`<ul style="padding-left:18px;margin:0;line-height:1.6;color:#374151;font-size:14px;">`)
		for _, h := range digest.Highlights {
			b.WriteString(`<li style="margin-bottom:4px;">`)
			b.WriteString(html.EscapeString(h))
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul></div>`)
	}

	// Collect action items from all emails
	var allActionItems []string
	for _, es := range digest.EmailSummaries {
		allActionItems = append(allActionItems, es.ActionItems...)
	}
	if len(allActionItems) > 0 {
		b.WriteString(`<div style="background:#fffbeb;border-left:3px solid #f59e0b;border-radius:0 8px 8px 0;padding:16px;margin-bottom:24px;">`)
		b.WriteString(`<h2 style="font-size:14px;font-weight:600;color:#92400e;margin:0 0 10px 0;">Action Items</h2>`)
		b.WriteString(`<ul style="padding-left:18px;margin:0;line-height:1.6;color:#78350f;font-size:14px;">`)
		for _, item := range allActionItems {
			b.WriteString(`<li style="margin-bottom:4px;">`)
			b.WriteString(html.EscapeString(item))
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul></div>`)
	}

	// Topics
	if len(digest.TopTopics) > 0 {
		b.WriteString(`<div style="margin-bottom:24px;">`)
		for _, topic := range digest.TopTopics {
			b.WriteString(`<span style="display:inline-block;background:#f1f5f9;color:#475569;border-radius:100px;padding:4px 12px;margin:2px 4px 2px 0;font-size:12px;font-weight:500;">`)
			b.WriteString(html.EscapeString(topic))
			b.WriteString(`</span>`)
		}
		b.WriteString(`</div>`)
	}

	// Divider before emails
	b.WriteString(`<hr style="border:none;border-top:1px solid #e5e7eb;margin:0 0 24px 0;">`)

	// Email summaries
	if len(digest.EmailSummaries) > 0 {
		b.WriteString(`<h2 style="font-size:15px;font-weight:600;color:#374151;margin:0 0 16px 0;">Email Summaries</h2>`)
		for _, es := range digest.EmailSummaries {
			b.WriteString(`<div style="margin-bottom:16px;padding:14px 16px;border:1px solid #e5e7eb;border-radius:8px;">`)

			// Subject line with category badge
			b.WriteString(`<div style="margin-bottom:6px;">`)
			b.WriteString(`<span style="font-weight:600;font-size:14px;color:#111827;">`)
			b.WriteString(html.EscapeString(es.Subject))
			b.WriteString(`</span>`)
			if es.Category != "" {
				fg := categoryColors[es.Category]
				bg := categoryBgColors[es.Category]
				if fg == "" {
					fg = "#6b7280"
				}
				if bg == "" {
					bg = "#f3f4f6"
				}
				b.WriteString(fmt.Sprintf(` <span style="display:inline-block;background:%s;color:%s;border-radius:100px;padding:1px 8px;font-size:11px;font-weight:500;vertical-align:middle;">`, bg, fg))
				b.WriteString(html.EscapeString(es.Category))
				b.WriteString(`</span>`)
			}
			b.WriteString(`</div>`)

			// Sender
			if es.SenderName != "" {
				b.WriteString(`<div style="font-size:12px;color:#6b7280;margin-bottom:8px;">`)
				b.WriteString(html.EscapeString(es.SenderName))
				b.WriteString(`</div>`)
			}

			// Summary
			b.WriteString(`<div style="font-size:13px;line-height:1.5;color:#374151;">`)
			b.WriteString(html.EscapeString(es.Summary))
			b.WriteString(`</div>`)

			// Per-email action items
			if len(es.ActionItems) > 0 {
				b.WriteString(`<div style="margin-top:8px;padding-top:8px;border-top:1px solid #f3f4f6;">`)
				b.WriteString(`<div style="font-size:11px;font-weight:600;color:#92400e;margin-bottom:4px;">Action items:</div>`)
				for _, ai := range es.ActionItems {
					b.WriteString(`<div style="font-size:12px;color:#78350f;line-height:1.4;padding-left:12px;">• `)
					b.WriteString(html.EscapeString(ai))
					b.WriteString(`</div>`)
				}
				b.WriteString(`</div>`)
			}

			// Link to dashboard
			b.WriteString(`<a href="https://mailbrief.io/inbox/`)
			b.WriteString(html.EscapeString(es.EmailID))
			b.WriteString(`" style="display:inline-block;margin-top:8px;font-size:12px;color:#0066cc;text-decoration:none;">View full email &rarr;</a>`)

			b.WriteString(`</div>`)
		}
	}

	// Footer
	b.WriteString(`<hr style="border:none;border-top:1px solid #e5e7eb;margin-top:32px;">`)
	b.WriteString(`<p style="color:#9ca3af;font-size:12px;line-height:1.5;">`)
	b.WriteString(`Sent by <a href="https://mailbrief.io" style="color:#9ca3af;">MailBrief</a>. `)
	b.WriteString(`Manage delivery in your <a href="https://mailbrief.io/settings" style="color:#9ca3af;">settings</a>`)
	if unsubscribeURL != "" {
		b.WriteString(` or <a href="`)
		b.WriteString(html.EscapeString(unsubscribeURL))
		b.WriteString(`" style="color:#9ca3af;">unsubscribe</a>`)
	}
	b.WriteString(`.`)
	b.WriteString(`</p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}
