package mailer

import (
	"strings"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
)

var (
	testPeriodStart = time.Date(2026, time.March, 4, 0, 0, 0, 0, time.UTC)
	testPeriodEnd   = time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
)

func TestBuildDigestHTMLDateRange(t *testing.T) {
	got := buildDigestHTML(&domain.Digest{
		Title:       "Digest",
		PeriodStart: testPeriodStart,
		PeriodEnd:   testPeriodEnd,
	}, "")

	if want := "2026/03/04 – 2026/03/10"; !strings.Contains(got, want) {
		t.Errorf("digest HTML missing date range %q", want)
	}
}

func TestBuildWeeklySummaryHTMLDateRange(t *testing.T) {
	got := buildWeeklySummaryHTML(&domain.WeeklySummary{
		PeriodStart: testPeriodStart,
		PeriodEnd:   testPeriodEnd,
	}, "")

	if want := "2026/03/04 – 2026/03/10"; !strings.Contains(got, want) {
		t.Errorf("weekly summary HTML missing date range %q", want)
	}
}

func TestWeeklySummarySubject(t *testing.T) {
	got := weeklySummarySubject(&domain.WeeklySummary{
		PeriodStart: testPeriodStart,
		PeriodEnd:   testPeriodEnd,
	})

	if want := "Your week in review — 4 March to 10 March"; got != want {
		t.Errorf("subject = %q, want %q", got, want)
	}
}

func TestBuildSignupNotificationHTML(t *testing.T) {
	body := buildSignupNotificationHTML(`evil<script>@example.com`, "open registration")
	if strings.Contains(body, "<script>") {
		t.Error("new user's email address was not HTML-escaped")
	}
	for _, want := range []string{"evil&lt;script&gt;@example.com", "via open registration.", "https://mailbrief.io/admin"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
	if got := signupNotificationSubject("new@example.com"); got != "MailBrief: New sign-up from new@example.com" {
		t.Errorf("subject = %q", got)
	}
}

func TestBuildUserFeedbackHTML(t *testing.T) {
	body := buildUserFeedbackHTML("reader@example.com", "line one\nline two <script>alert(1)</script>")

	if strings.Contains(body, "<script>") {
		t.Error("the message was not HTML-escaped")
	}
	for _, want := range []string{"reader@example.com", "white-space:pre-wrap", "&lt;script&gt;"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
