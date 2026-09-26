package database

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestListForUsersMatchesGet is the point of the batch path: it must return
// exactly what Get returns, including the defaults for a user with no row.
// Two implementations of the same defaults would drift silently.
func TestListForUsersMatchesGet(t *testing.T) {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	r := &PreferenceRepo{pool: pool}

	// "saved" has a row; "unsaved" deliberately does not.
	ids := []string{"pref-saved", "pref-unsaved"}

	batch, err := r.ListForUsers(ctx, ids)
	if err != nil {
		t.Fatalf("ListForUsers: %v", err)
	}
	if len(batch) != len(ids) {
		t.Fatalf("got %d preferences, want %d — a user with no row must still get defaults", len(batch), len(ids))
	}

	for _, id := range ids {
		one, err := r.Get(ctx, id)
		if err != nil {
			t.Fatalf("Get(%s): %v", id, err)
		}
		b := batch[id]
		if b == nil {
			t.Fatalf("ListForUsers returned nothing for %s", id)
		}
		if b.DigestFrequency != one.DigestFrequency ||
			b.DigestHour != one.DigestHour ||
			b.DigestDay != one.DigestDay ||
			b.DigestTimezone != one.DigestTimezone ||
			b.DigestStyle != one.DigestStyle ||
			b.TopicLimit != one.TopicLimit ||
			b.MonthlyTokenLimit != one.MonthlyTokenLimit ||
			b.DigestEmail != one.DigestEmail ||
			b.WeeklySummary != one.WeeklySummary ||
			len(b.ExcludedCategories) != len(one.ExcludedCategories) {
			t.Errorf("%s: batch and Get disagree\n batch: %+v\n get:   %+v", id, b, one)
		}
	}

	if _, err := r.ListForUsers(ctx, nil); err != nil {
		t.Errorf("ListForUsers(nil): %v", err)
	}
}
