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

func (m *ResendMailer) SendInvite(ctx context.Context, toEmail string) error {
	return m.sendInviteVariant(ctx, toEmail,
		"You're invited to MailBrief",
		"You've been invited to join MailBrief — AI-powered digests of your newsletters, delivered daily.")
}

func (m *ResendMailer) SendInviteApproved(ctx context.Context, toEmail string) error {
	return m.sendInviteVariant(ctx, toEmail,
		"Your MailBrief access has been approved",
		"Your request to join MailBrief has been approved. You can now sign in and start receiving AI-powered digests of your newsletters.")
}

func (m *ResendMailer) sendInviteVariant(ctx context.Context, toEmail, subject, lead string) error {
	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: subject,
		Html:    buildInviteHTML(subject, lead),
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send invite email: %w", err)
	}
	return nil
}

func buildInviteHTML(title, lead string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#1a1a1a;">`)

	b.WriteString(`<h1 style="font-size:24px;margin-bottom:8px;">`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</h1>`)

	b.WriteString(`<p style="font-size:16px;line-height:1.5;">`)
	b.WriteString(html.EscapeString(lead))
	b.WriteString(`</p>`)

	b.WriteString(`<p style="font-size:16px;line-height:1.5;">Sign in to get started:</p>`)

	b.WriteString(`<a href="https://mailbrief.io/login" style="display:inline-block;background:#0066cc;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-size:16px;font-weight:500;">Sign in to MailBrief</a>`)

	b.WriteString(`<hr style="border:none;border-top:1px solid #eee;margin-top:32px;">`)
	b.WriteString(`<p style="color:#999;font-size:12px;">Sent by <a href="https://mailbrief.io" style="color:#999;">MailBrief</a></p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}

func (m *ResendMailer) SendTokenWarning(ctx context.Context, toEmail string, usagePercent int, tokensUsed, tokenLimit int) error {
	htmlBody := buildTokenWarningHTML(usagePercent, tokensUsed, tokenLimit)

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("MailBrief: You've used %d%% of your monthly token limit", usagePercent),
		Html:    htmlBody,
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send token warning email: %w", err)
	}
	return nil
}

func (m *ResendMailer) SendWaitlistNotification(ctx context.Context, toEmail, applicantEmail string) error {
	htmlBody := buildWaitlistNotificationHTML(applicantEmail)

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("MailBrief: New waitlist application from %s", applicantEmail),
		Html:    htmlBody,
	}

	_, err := m.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send waitlist notification: %w", err)
	}
	return nil
}

func buildWaitlistNotificationHTML(applicantEmail string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#1a1a1a;">`)

	b.WriteString(`<h1 style="font-size:22px;margin-bottom:8px;">New Waitlist Application</h1>`)

	b.WriteString(`<p style="font-size:16px;line-height:1.5;"><strong>`)
	b.WriteString(html.EscapeString(applicantEmail))
	b.WriteString(`</strong> has requested access to MailBrief.</p>`)

	b.WriteString(`<a href="https://mailbrief.io/admin" style="display:inline-block;background:#0066cc;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-size:16px;font-weight:500;">Review in Admin</a>`)

	b.WriteString(`<hr style="border:none;border-top:1px solid #eee;margin-top:32px;">`)
	b.WriteString(`<p style="color:#999;font-size:12px;">Sent by <a href="https://mailbrief.io" style="color:#999;">MailBrief</a></p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}

func buildTokenWarningHTML(usagePercent int, tokensUsed, tokenLimit int) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#1a1a1a;">`)

	b.WriteString(`<h1 style="font-size:22px;margin-bottom:8px;">Token Usage Alert</h1>`)

	b.WriteString(fmt.Sprintf(`<p style="font-size:16px;line-height:1.5;">You've used <strong>%d%%</strong> of your monthly token limit on MailBrief.</p>`, usagePercent))

	// Usage bar
	b.WriteString(`<div style="background:#f3f4f6;border-radius:8px;height:24px;margin:16px 0;overflow:hidden;">`)
	barColor := "#f59e0b"
	if usagePercent >= 95 {
		barColor = "#ef4444"
	}
	b.WriteString(fmt.Sprintf(`<div style="background:%s;height:100%%;width:%d%%;border-radius:8px;"></div>`, barColor, usagePercent))
	b.WriteString(`</div>`)

	b.WriteString(fmt.Sprintf(`<p style="font-size:14px;color:#6b7280;">%s / %s tokens used this month.</p>`,
		formatTokenCount(tokensUsed), formatTokenCount(tokenLimit)))

	b.WriteString(`<p style="font-size:14px;line-height:1.5;color:#374151;">Once you reach your limit, email processing will pause until the next billing period. You can increase your limit or add your own API key in settings.</p>`)

	b.WriteString(`<a href="https://mailbrief.io/settings" style="display:inline-block;background:#0066cc;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;font-size:16px;font-weight:500;margin-top:8px;">Manage Settings</a>`)

	b.WriteString(`<hr style="border:none;border-top:1px solid #eee;margin-top:32px;">`)
	b.WriteString(`<p style="color:#999;font-size:12px;">Sent by <a href="https://mailbrief.io" style="color:#999;">MailBrief</a></p>`)

	b.WriteString(`</body></html>`)

	return b.String()
}

func formatTokenCount(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.0fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
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

func (m *ResendMailer) SendWeeklySummary(ctx context.Context, toEmail string, summary *domain.WeeklySummary, unsubscribeURL string) error {
	htmlBody := buildWeeklySummaryHTML(summary, unsubscribeURL)

	subject := fmt.Sprintf("Your week in review — %s to %s",
		summary.PeriodStart.Format("Jan 2"),
		summary.PeriodEnd.Format("Jan 2"))

	params := &resend.SendEmailRequest{
		From:    m.fromAddress,
		To:      []string{toEmail},
		Subject: subject,
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
		return fmt.Errorf("send weekly summary: %w", err)
	}
	return nil
}

func buildWeeklySummaryHTML(summary *domain.WeeklySummary, unsubscribeURL string) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>`)
	b.WriteString(`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:24px 16px;color:#1a1a1a;background:#ffffff;">`)

	// Header
	b.WriteString(`<table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:24px;"><tr>`)
	b.WriteString(`<td><span style="font-size:13px;font-weight:600;color:#6b7280;letter-spacing:0.05em;text-transform:uppercase;">MailBrief</span></td>`)
	b.WriteString(`<td align="right"><span style="font-size:13px;color:#6b7280;">`)
	b.WriteString(summary.PeriodStart.Format("Jan 2"))
	b.WriteString(` – `)
	b.WriteString(summary.PeriodEnd.Format("Jan 2, 2006"))
	b.WriteString(`</span></td></tr></table>`)

	b.WriteString(`<h1 style="font-size:22px;font-weight:700;margin:0 0 20px 0;line-height:1.3;">Your week in review</h1>`)

	// Stats
	b.WriteString(`<div style="background:#f8fafc;border-radius:8px;padding:20px;margin-bottom:24px;">`)
	b.WriteString(fmt.Sprintf(`<div style="font-size:32px;font-weight:700;color:#111827;">%d</div>`, summary.EmailCount))
	b.WriteString(`<div style="font-size:14px;color:#6b7280;margin-top:2px;">emails received</div>`)
	if summary.ActionItems > 0 {
		b.WriteString(fmt.Sprintf(`<div style="font-size:14px;color:#d97706;margin-top:8px;font-weight:500;">%d action items identified</div>`, summary.ActionItems))
	}
	b.WriteString(`</div>`)

	// Top senders
	if len(summary.TopSenders) > 0 {
		b.WriteString(`<h2 style="font-size:14px;font-weight:600;color:#374151;margin:0 0 12px 0;text-transform:uppercase;letter-spacing:0.03em;">Top senders</h2>`)
		b.WriteString(`<table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:24px;">`)
		for _, s := range summary.TopSenders {
			name := s.FromName
			if name == "" {
				name = s.FromAddress
			}
			b.WriteString(`<tr><td style="padding:8px 0;border-bottom:1px solid #f3f4f6;font-size:14px;color:#374151;">`)
			b.WriteString(html.EscapeString(name))
			b.WriteString(`</td><td align="right" style="padding:8px 0;border-bottom:1px solid #f3f4f6;font-size:14px;color:#6b7280;">`)
			b.WriteString(fmt.Sprintf("%d", s.Count))
			b.WriteString(`</td></tr>`)
		}
		b.WriteString(`</table>`)
	}

	// Categories
	if len(summary.Categories) > 0 {
		b.WriteString(`<h2 style="font-size:14px;font-weight:600;color:#374151;margin:0 0 12px 0;text-transform:uppercase;letter-spacing:0.03em;">By category</h2>`)
		b.WriteString(`<table width="100%" cellpadding="0" cellspacing="0" style="margin-bottom:24px;">`)
		for _, cat := range []string{"newsletter", "personal", "transactional", "marketing", "notification", "other"} {
			count, ok := summary.Categories[cat]
			if !ok || count == 0 {
				continue
			}
			color := categoryColors[cat]
			if color == "" {
				color = "#6b7280"
			}
			b.WriteString(`<tr><td style="padding:8px 0;border-bottom:1px solid #f3f4f6;font-size:14px;">`)
			b.WriteString(fmt.Sprintf(`<span style="display:inline-block;width:8px;height:8px;border-radius:50%%;background:%s;margin-right:8px;"></span>`, color))
			b.WriteString(html.EscapeString(strings.ToUpper(cat[:1]) + cat[1:]))
			b.WriteString(`</td><td align="right" style="padding:8px 0;border-bottom:1px solid #f3f4f6;font-size:14px;color:#6b7280;">`)
			b.WriteString(fmt.Sprintf("%d", count))
			b.WriteString(`</td></tr>`)
		}
		b.WriteString(`</table>`)
	}

	// CTA
	b.WriteString(`<div style="text-align:center;margin:32px 0;">`)
	b.WriteString(`<a href="https://mailbrief.io/inbox" style="display:inline-block;background:#111827;color:#ffffff;padding:12px 28px;border-radius:6px;text-decoration:none;font-size:15px;font-weight:500;">View your inbox</a>`)
	b.WriteString(`</div>`)

	// Footer
	b.WriteString(`<hr style="border:none;border-top:1px solid #e5e7eb;margin:32px 0 16px 0;">`)
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
