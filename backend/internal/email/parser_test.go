package email

import (
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestParse_SimpleTextEmail(t *testing.T) {
	raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:      "sender@example.com",
		To:        "recipient@example.com",
		Subject:   "Test Subject",
		MessageID: "<abc123@example.com>",
		TextBody:  "Hello, this is the body.",
	})

	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.From != "sender@example.com" {
		t.Errorf("From = %q, want %q", parsed.From, "sender@example.com")
	}
	if parsed.To != "recipient@example.com" {
		t.Errorf("To = %q, want %q", parsed.To, "recipient@example.com")
	}
	if parsed.Subject != "Test Subject" {
		t.Errorf("Subject = %q, want %q", parsed.Subject, "Test Subject")
	}
	if parsed.MessageID != "<abc123@example.com>" {
		t.Errorf("MessageID = %q, want %q", parsed.MessageID, "<abc123@example.com>")
	}
	if parsed.TextBody != "Hello, this is the body." {
		t.Errorf("TextBody = %q, want %q", parsed.TextBody, "Hello, this is the body.")
	}
}

func TestParse_MultipartAlternative(t *testing.T) {
	raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "sender@example.com",
		To:       "recipient@example.com",
		Subject:  "Multipart",
		TextBody: "Plain text version",
		HTMLBody: "<p>HTML version</p>",
	})

	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.TextBody == "" {
		t.Error("expected non-empty TextBody")
	}
	if parsed.HTMLBody == "" {
		t.Error("expected non-empty HTMLBody")
	}
}

func TestParse_FromNameExtraction(t *testing.T) {
	raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "Jane Doe <jane@example.com>",
		To:       "recipient@example.com",
		Subject:  "Test",
		TextBody: "body",
	})

	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.From != "jane@example.com" {
		t.Errorf("From = %q, want %q", parsed.From, "jane@example.com")
	}
	if parsed.FromName != "Jane Doe" {
		t.Errorf("FromName = %q, want %q", parsed.FromName, "Jane Doe")
	}
}

func TestParse_DateHeaderFormats(t *testing.T) {
	tests := []struct {
		name string
		date string
	}{
		{"RFC1123Z", "Mon, 02 Jan 2006 15:04:05 -0700"},
		{"RFC822Z", "02 Jan 06 15:04 -0700"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
				From:     "sender@example.com",
				To:       "recipient@example.com",
				Subject:  "Test",
				Date:     tt.date,
				TextBody: "body",
			})

			parsed, err := Parse(raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.DateHeader.IsZero() {
				t.Error("expected non-zero DateHeader")
			}
		})
	}
}

func TestParse_MissingDateHeader(t *testing.T) {
	raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "sender@example.com",
		To:       "recipient@example.com",
		Subject:  "No Date",
		TextBody: "body",
	})

	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !parsed.DateHeader.Equal(time.Time{}) {
		t.Errorf("DateHeader = %v, want zero time", parsed.DateHeader)
	}
}

func TestParse_EmptyBody(t *testing.T) {
	raw := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:    "sender@example.com",
		To:      "recipient@example.com",
		Subject: "Empty",
	})

	parsed, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.TextBody != "" {
		t.Errorf("TextBody = %q, want empty", parsed.TextBody)
	}
	if parsed.HTMLBody != "" {
		t.Errorf("HTMLBody = %q, want empty", parsed.HTMLBody)
	}
}

func TestParse_InvalidMIME(t *testing.T) {
	_, err := Parse([]byte("this is not a valid email"))
	if err == nil {
		t.Error("expected error for invalid MIME, got nil")
	}
}
