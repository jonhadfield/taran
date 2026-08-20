package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/crypto"
)

// newEmailBackfillTestRepo connects to TARAN_TEST_DB_URL. The backend CI job
// runs without Postgres, so these skip unless a database is supplied:
//
//	TARAN_TEST_DB_URL=postgresql://... go test ./internal/database/
func newEmailBackfillTestRepo(t *testing.T) (*EmailRepo, context.Context) {
	t.Helper()

	dbURL := os.Getenv("TARAN_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TARAN_TEST_DB_URL not set; skipping database-backed test")
	}

	ctx := context.Background()
	pool, err := NewPool(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	for _, stmt := range []string{
		`DELETE FROM email`,
		`DELETE FROM email_account`,
		`DELETE FROM "user"`,
		`INSERT INTO "user"(id, email) VALUES ('bf-user', 'bf@test')`,
		`INSERT INTO email_account(id, user_id, email_address) VALUES ('bf-acct', 'bf-user', 'bf@mail.test')`,
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup %q: %v", stmt, err)
		}
	}

	encryptor, err := crypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	return NewEmailRepo(pool, encryptor), ctx
}

// insertPlaintextEmail writes a row the way the pre-encryption code did.
func insertPlaintextEmail(t *testing.T, r *EmailRepo, ctx context.Context, id, text, html string) {
	t.Helper()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email (id, user_id, email_account_id, message_id, from_address, to_address,
		                    subject, text_body, html_body, received_at, date_header, encrypted)
		 VALUES ($1, 'bf-user', 'bf-acct', $1, 'sender@x', 'bf@mail.test', 'Subject', $2, $3, $4, $4, FALSE)`,
		id, text, html, time.Now())
	if err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

func TestEmailBackfillEncryption(t *testing.T) {
	repo, ctx := newEmailBackfillTestRepo(t)

	insertPlaintextEmail(t, repo, ctx, "e1", "plain text one", "<p>html one</p>")
	insertPlaintextEmail(t, repo, ctx, "e2", "plain text two", "")
	insertPlaintextEmail(t, repo, ctx, "e3", "", "<p>html only</p>")
	insertPlaintextEmail(t, repo, ctx, "e4", "", "") // no bodies at all

	t.Run("dry run changes nothing", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, true, nil)
		if err != nil {
			t.Fatalf("dry run: %v", err)
		}
		if stats.Encrypted != 3 || stats.SkippedEmpty != 1 {
			t.Errorf("stats = %+v, want 3 encrypted / 1 empty", stats)
		}
		remaining, err := repo.CountUnencrypted(ctx)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if remaining != 4 {
			t.Errorf("after dry run %d rows remain plaintext, want 4", remaining)
		}
	})

	t.Run("encrypts and bodies survive the round trip", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, false, nil)
		if err != nil {
			t.Fatalf("backfill: %v", err)
		}
		if stats.Failed != 0 {
			t.Fatalf("stats = %+v, want 0 failed", stats)
		}

		for _, tc := range []struct{ id, text, html string }{
			{"e1", "plain text one", "<p>html one</p>"},
			{"e2", "plain text two", ""},
			{"e3", "", "<p>html only</p>"},
			{"e4", "", ""},
		} {
			got, err := repo.GetByID(ctx, "bf-user", tc.id)
			if err != nil {
				t.Fatalf("read %s: %v", tc.id, err)
			}
			if got.TextBody != tc.text {
				t.Errorf("%s text = %q, want %q", tc.id, got.TextBody, tc.text)
			}
			if got.HTMLBody != tc.html {
				t.Errorf("%s html = %q, want %q", tc.id, got.HTMLBody, tc.html)
			}
		}

		// The columns themselves must no longer hold the plaintext.
		var storedText string
		if err := repo.pool.QueryRow(ctx,
			`SELECT text_body FROM email WHERE id = 'e1'`).Scan(&storedText); err != nil {
			t.Fatalf("read raw column: %v", err)
		}
		if storedText == "plain text one" {
			t.Error("text_body still contains plaintext")
		}
	})

	t.Run("marks every row so re-runs converge", func(t *testing.T) {
		remaining, err := repo.CountUnencrypted(ctx)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if remaining != 0 {
			t.Errorf("%d rows still flagged unencrypted, want 0", remaining)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, false, nil)
		if err != nil {
			t.Fatalf("second run: %v", err)
		}
		if stats.Scanned != 0 {
			t.Errorf("second run scanned %d rows, want 0", stats.Scanned)
		}
	})
}

func TestEmailBackfillEncryption_RequiresEncryptor(t *testing.T) {
	repo := &EmailRepo{}
	if _, err := repo.BackfillEncryption(context.Background(), 10, false, nil); err == nil {
		t.Error("expected an error when no encryptor is configured")
	}
}

func TestEmailBackfillEncryption_PaginatesPastMultipleBatches(t *testing.T) {
	repo, ctx := newEmailBackfillTestRepo(t)

	for _, id := range []string{"b1", "b2", "b3", "b4", "b5"} {
		insertPlaintextEmail(t, repo, ctx, id, "body "+id, "")
	}

	stats, err := repo.BackfillEncryption(ctx, 2, false, nil)
	if err != nil {
		t.Fatalf("backfill: %v", err)
	}
	if stats.Encrypted != 5 {
		t.Errorf("encrypted %d rows, want 5", stats.Encrypted)
	}
	remaining, err := repo.CountUnencrypted(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if remaining != 0 {
		t.Errorf("%d rows remain plaintext, want 0", remaining)
	}
}
