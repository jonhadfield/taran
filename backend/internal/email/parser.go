package email

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jhillyerd/enmime"
)

type Attachment struct {
	Filename    string
	ContentType string
	SizeBytes   int
}

type ParsedEmail struct {
	MessageID        string
	From             string
	FromName         string
	To               string
	Subject          string
	TextBody         string
	HTMLBody         string
	DateHeader       time.Time
	Attachments      []Attachment
	UnsubscribeURL   string // HTTP(S) URL from List-Unsubscribe header
	UnsubscribeMailto string // mailto address from List-Unsubscribe header
	UnsubscribePost  bool   // true if List-Unsubscribe-Post header present (RFC 8058)
}

func Parse(raw []byte) (*ParsedEmail, error) {
	env, err := enmime.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parse email: %w", err)
	}

	var dateHeader time.Time
	if d := env.GetHeader("Date"); d != "" {
		dateHeader, _ = parseDate(d)
	}

	from := env.GetHeader("From")
	fromName := ""
	if addrs, err := env.AddressList("From"); err == nil && len(addrs) > 0 {
		from = addrs[0].Address
		fromName = addrs[0].Name
	}

	to := env.GetHeader("To")
	if addrs, err := env.AddressList("To"); err == nil && len(addrs) > 0 {
		to = addrs[0].Address
	}

	var attachments []Attachment
	for _, a := range env.Attachments {
		attachments = append(attachments, Attachment{
			Filename:    a.FileName,
			ContentType: a.ContentType,
			SizeBytes:   len(a.Content),
		})
	}

	unsubURL, unsubMailto := parseListUnsubscribe(env.GetHeader("List-Unsubscribe"))
	unsubPost := strings.Contains(env.GetHeader("List-Unsubscribe-Post"), "List-Unsubscribe=One-Click")

	return &ParsedEmail{
		MessageID:         env.GetHeader("Message-Id"),
		From:              from,
		FromName:          fromName,
		To:                to,
		Subject:           env.GetHeader("Subject"),
		TextBody:          env.Text,
		HTMLBody:          env.HTML,
		DateHeader:        dateHeader,
		Attachments:       attachments,
		UnsubscribeURL:    unsubURL,
		UnsubscribeMailto: unsubMailto,
		UnsubscribePost:   unsubPost,
	}, nil
}

// parseListUnsubscribe extracts HTTP and mailto URLs from a List-Unsubscribe header.
// Format: <https://example.com/unsub>, <mailto:unsub@example.com>
func parseListUnsubscribe(header string) (httpURL, mailto string) {
	if header == "" {
		return "", ""
	}
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		// Strip angle brackets
		part = strings.TrimPrefix(part, "<")
		part = strings.TrimSuffix(part, ">")
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "https://") || strings.HasPrefix(part, "http://") {
			if httpURL == "" {
				httpURL = part
			}
		} else if strings.HasPrefix(part, "mailto:") {
			if mailto == "" {
				mailto = strings.TrimPrefix(part, "mailto:")
			}
		}
	}
	return httpURL, mailto
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05 -0700",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
