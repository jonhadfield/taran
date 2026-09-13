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
