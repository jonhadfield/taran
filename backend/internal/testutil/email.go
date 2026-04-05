package testutil

import (
	"fmt"
	"strings"
)

// RawEmailOpts configures BuildRawEmail output.
type RawEmailOpts struct {
	From      string // "Name <addr>" or just "addr"
	To        string
	Subject   string
	Date      string // RFC 1123Z format; empty = omit
	MessageID string // empty = omit
	TextBody  string
	HTMLBody  string // if set, produces multipart/alternative
}

// BuildRawEmail constructs a valid RFC 2822 MIME message from the given options.
func BuildRawEmail(opts RawEmailOpts) []byte {
	if opts.HTMLBody != "" {
		return buildMultipartEmail(opts)
	}
	return buildTextEmail(opts)
}

func buildTextEmail(opts RawEmailOpts) []byte {
	var b strings.Builder
	writeHeaders(&b, opts)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(opts.TextBody)
	return []byte(b.String())
}

func buildMultipartEmail(opts RawEmailOpts) []byte {
	boundary := "test-boundary-12345"
	var b strings.Builder
	writeHeaders(&b, opts)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	b.WriteString("\r\n")

	if opts.TextBody != "" {
		b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
		b.WriteString("\r\n")
		b.WriteString(opts.TextBody)
		b.WriteString("\r\n")
	}

	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(opts.HTMLBody)
	b.WriteString("\r\n")

	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return []byte(b.String())
}

func writeHeaders(b *strings.Builder, opts RawEmailOpts) {
	if opts.From != "" {
		b.WriteString(fmt.Sprintf("From: %s\r\n", opts.From))
	}
	if opts.To != "" {
		b.WriteString(fmt.Sprintf("To: %s\r\n", opts.To))
	}
	if opts.Subject != "" {
		b.WriteString(fmt.Sprintf("Subject: %s\r\n", opts.Subject))
	}
	if opts.Date != "" {
		b.WriteString(fmt.Sprintf("Date: %s\r\n", opts.Date))
	}
	if opts.MessageID != "" {
		b.WriteString(fmt.Sprintf("Message-Id: %s\r\n", opts.MessageID))
	}
}
