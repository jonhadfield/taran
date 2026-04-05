package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestStatsHandler_Get_Success(t *testing.T) {
	h := &StatsHandler{
		Emails: &testutil.MockEmailRepo{
			CountByPeriodFn: func(_ context.Context, _ string, from, to time.Time) (int, error) {
				// Distinguish this/last week by checking the range width
				if to.Sub(from) > 7*24*time.Hour {
					return 0, nil // won't match; both calls are <=7 days
				}
				// The first call is this week, second is last week.
				// We differentiate by checking if "to" is recent.
				if to.After(time.Now().Add(-time.Minute)) {
					return 12, nil // this week
				}
				return 8, nil // last week
			},
			CountByFilterFn: func(_ context.Context, _ string, _ *domain.EmailStatus, _ *bool) (int, error) {
				return 42, nil
			},
			TopSendersFn: func(_ context.Context, _ string, _, _ time.Time, _ int) ([]domain.SenderCount, error) {
				return []domain.SenderCount{
					{FromAddress: "news@example.com", FromName: "News", Count: 5},
				}, nil
			},
		},
	}

	req := httptest.NewRequest("GET", "/api/stats", nil)
	req = req.WithContext(testutil.ContextWithUserID("user-1"))
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var stats domain.UserStats
	json.NewDecoder(rec.Body).Decode(&stats)
	if stats.EmailsThisWeek != 12 {
		t.Errorf("EmailsThisWeek = %d, want 12", stats.EmailsThisWeek)
	}
	if stats.EmailsLastWeek != 8 {
		t.Errorf("EmailsLastWeek = %d, want 8", stats.EmailsLastWeek)
	}
	if stats.TotalEmails != 42 {
		t.Errorf("TotalEmails = %d, want 42", stats.TotalEmails)
	}
	if len(stats.TopSenders) != 1 {
		t.Fatalf("TopSenders len = %d, want 1", len(stats.TopSenders))
	}
}

func TestStartOfWeek(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "monday stays monday",
			in:   time.Date(2025, 2, 17, 14, 30, 0, 0, time.UTC), // Monday
			want: time.Date(2025, 2, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "wednesday goes to monday",
			in:   time.Date(2025, 2, 19, 10, 0, 0, 0, time.UTC), // Wednesday
			want: time.Date(2025, 2, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "sunday goes to monday of same week",
			in:   time.Date(2025, 2, 23, 23, 59, 0, 0, time.UTC), // Sunday
			want: time.Date(2025, 2, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "saturday goes to monday",
			in:   time.Date(2025, 2, 22, 8, 0, 0, 0, time.UTC), // Saturday
			want: time.Date(2025, 2, 17, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := startOfWeek(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("startOfWeek(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
