package digest

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/llm"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestShouldGenerateForUser_MatchesHour(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "daily",
		DigestHour:      8,
		DigestTimezone:  "America/New_York",
	}

	// 8am ET = 13:00 UTC (EST, no DST)
	nowUTC := time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC) // Thursday
	if !shouldGenerateForUser(pref, nowUTC) {
		t.Error("expected shouldGenerate=true when UTC hour maps to user's preferred hour")
	}
}

func TestShouldGenerateForUser_WrongHour(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "daily",
		DigestHour:      8,
		DigestTimezone:  "America/New_York",
	}

	nowUTC := time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC) // 10am ET, not 8am
	if shouldGenerateForUser(pref, nowUTC) {
		t.Error("expected shouldGenerate=false when hour doesn't match")
	}
}

func TestShouldGenerateForUser_WeeklyOnMonday(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "weekly",
		DigestHour:      9,
		DigestDay:       1, // Monday
		DigestTimezone:  "UTC",
	}

	// Monday 9am UTC
	monday := time.Date(2026, 2, 16, 9, 0, 0, 0, time.UTC)
	if monday.Weekday() != time.Monday {
		t.Fatalf("expected Monday, got %s", monday.Weekday())
	}
	if !shouldGenerateForUser(pref, monday) {
		t.Error("expected shouldGenerate=true for weekly on Monday at correct hour")
	}
}

func TestShouldGenerateForUser_WeeklyNotMonday(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "weekly",
		DigestHour:      9,
		DigestDay:       1, // Monday
		DigestTimezone:  "UTC",
	}

	// Wednesday 9am UTC
	wednesday := time.Date(2026, 2, 18, 9, 0, 0, 0, time.UTC)
	if wednesday.Weekday() != time.Wednesday {
		t.Fatalf("expected Wednesday, got %s", wednesday.Weekday())
	}
	if shouldGenerateForUser(pref, wednesday) {
		t.Error("expected shouldGenerate=false for weekly on non-Monday")
	}
}

func TestShouldGenerateForUser_InvalidTimezone_FallsBackToUTC(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "daily",
		DigestHour:      10,
		DigestTimezone:  "Invalid/Zone",
	}

	nowUTC := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	if !shouldGenerateForUser(pref, nowUTC) {
		t.Error("expected shouldGenerate=true with fallback to UTC")
	}
}

func TestShouldGenerateForUser_WeeklyCustomDay(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "weekly",
		DigestHour:      9,
		DigestDay:       3, // Wednesday
		DigestTimezone:  "UTC",
	}

	// Wednesday 9am UTC
	wednesday := time.Date(2026, 2, 18, 9, 0, 0, 0, time.UTC)
	if wednesday.Weekday() != time.Wednesday {
		t.Fatalf("expected Wednesday, got %s", wednesday.Weekday())
	}
	if !shouldGenerateForUser(pref, wednesday) {
		t.Error("expected shouldGenerate=true for weekly on Wednesday when DigestDay=3")
	}

	// Monday 9am UTC — wrong day
	monday := time.Date(2026, 2, 16, 9, 0, 0, 0, time.UTC)
	if shouldGenerateForUser(pref, monday) {
		t.Error("expected shouldGenerate=false for weekly on Monday when DigestDay=3")
	}
}

func TestComputePeriod_Daily(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "daily",
		DigestHour:      8,
		DigestTimezone:  "UTC",
	}

	nowUTC := time.Date(2026, 2, 15, 8, 30, 0, 0, time.UTC)
	start, end := computePeriod(pref, nowUTC)

	expectedEnd := time.Date(2026, 2, 15, 8, 0, 0, 0, time.UTC)
	expectedStart := expectedEnd.Add(-24 * time.Hour)

	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
}

func TestComputePeriod_Weekly(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "weekly",
		DigestHour:      9,
		DigestTimezone:  "UTC",
	}

	nowUTC := time.Date(2026, 2, 16, 9, 15, 0, 0, time.UTC) // Monday
	start, end := computePeriod(pref, nowUTC)

	expectedEnd := time.Date(2026, 2, 16, 9, 0, 0, 0, time.UTC)
	expectedStart := expectedEnd.Add(-7 * 24 * time.Hour)

	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
}

func TestComputePeriod_WithTimezone(t *testing.T) {
	pref := &domain.UserPreference{
		DigestFrequency: "daily",
		DigestHour:      8,
		DigestTimezone:  "America/New_York",
	}

	// 13:00 UTC = 8:00 AM ET (EST)
	nowUTC := time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC)
	start, end := computePeriod(pref, nowUTC)

	// Period end should be 8:00 ET = 13:00 UTC
	expectedEnd := time.Date(2026, 1, 15, 13, 0, 0, 0, time.UTC)
	expectedStart := expectedEnd.Add(-24 * time.Hour)

	if !end.Equal(expectedEnd) {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
	if !start.Equal(expectedStart) {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
}

func TestRunNow_ConcurrentCallsOnlyOneExecutes(t *testing.T) {
	var generateCalls atomic.Int32

	s := NewScheduler(SchedulerConfig{
		Generator: &Generator{
			Extractions: &testutil.MockExtractionRepo{},
			Digests:     &testutil.MockDigestRepo{},
			Resolver:    llm.NewProviderResolver(&testutil.MockProvider{}, nil, nil, nil),
		},
		Emails: &testutil.MockEmailRepo{
			ListActiveUserIDsFn: func(_ context.Context, _, _ time.Time) ([]string, error) {
				// Simulate slow work so concurrent calls overlap
				time.Sleep(50 * time.Millisecond)
				generateCalls.Add(1)
				return nil, nil
			},
		},
		Digests:     &testutil.MockDigestRepo{},
		Preferences: &testutil.MockPreferenceRepo{},
		Sessions:    &testutil.MockSessionRepo{},
	})

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.RunNow()
		}()
	}
	wg.Wait()

	if got := generateCalls.Load(); got != 1 {
		t.Errorf("expected exactly 1 generation cycle, got %d", got)
	}
}
