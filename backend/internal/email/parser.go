package email

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jhillyerd/enmime"
)

type ParsedEmail struct {
	MessageID   string
	From        string
	FromName    string
	To          string
	Subject     string
	TextBody    string
	HTMLBody    string
	DateHeader  time.Time
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

	return &ParsedEmail{
		MessageID:  env.GetHeader("Message-Id"),
		From:       from,
		FromName:   fromName,
		To:         to,
		Subject:    env.GetHeader("Subject"),
		TextBody:   env.Text,
		HTMLBody:   env.HTML,
		DateHeader: dateHeader,
	}, nil
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
