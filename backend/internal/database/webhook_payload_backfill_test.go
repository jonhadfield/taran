package database

import (
	"context"
	"os"
	"testing"

	"github.com/hadfielj/taran/backend/internal/crypto"
)

// testKey is a throwaway 32-byte key; the backfill only needs a valid one.
const testKey = "4f8a2b7c1d3e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8"

// newBackfillTestRepo connects to the database named by TARAN_TEST_DB_URL.
// The backend CI job runs without Postgres, so these tests skip unless a
// database is supplied:
//
//	TARAN_TEST_DB_URL=postgresql://... go test ./internal/database/
func newBackfillTestRepo(t *testing.T) (*WebhookPayloadRepo, context.Context) {
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

	if _, err := pool.Exec(ctx, `DELETE FROM webhook_payload`); err != nil {
		t.Fatalf("clear webhook_payload: %v", err)
	}

	encryptor, err := crypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	return NewWebhookPayloadRepo(pool, encryptor), ctx
}

// insertPlaintext writes a row the way the pre-encryption code did.
func insertPlaintext(t *testing.T, r *WebhookPayloadRepo, ctx context.Context, id, body string) {
	t.Helper()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_payload (id, email_id, raw_body, headers, received_at, size_bytes, encrypted)
		 VALUES ($1, NULL, $2, '{}', NOW(), $3, FALSE)`, id, []byte(body), len(body))
	if err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

func TestBackfillEncryption(t *testing.T) {
	repo, ctx := newBackfillTestRepo(t)

	insertPlaintext(t, repo, ctx, "p1", "From: a@x\nSubject: Secret One\n\nbody one")
	insertPlaintext(t, repo, ctx, "p2", "From: b@x\nSubject: Secret Two\n\nbody two")
	insertPlaintext(t, repo, ctx, "p3", "")

	t.Run("dry run changes nothing", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, true, nil)
		if err != nil {
			t.Fatalf("dry run: %v", err)
		}
		if stats.Encrypted != 2 || stats.SkippedEmpty != 1 || stats.Failed != 0 {
			t.Errorf("stats = %+v, want 2 encrypted / 1 empty / 0 failed", stats)
		}
		remaining, err := repo.CountUnencrypted(ctx)
		if err != nil {
			t.Fatalf("count: %v", err)
		}
		if remaining != 2 {
			t.Errorf("after dry run %d rows remain plaintext, want 2", remaining)
		}
	})

	t.Run("encrypts and stays readable", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, false, nil)
		if err != nil {
			t.Fatalf("backfill: %v", err)
		}
		if stats.Encrypted != 2 || stats.Failed != 0 {
			t.Errorf("stats = %+v, want 2 encrypted / 0 failed", stats)
		}

		got, err := repo.GetByID(ctx, "p1")
		if err != nil {
			t.Fatalf("read back p1: %v", err)
		}
		if want := "From: a@x\nSubject: Secret One\n\nbody one"; string(got.RawBody) != want {
			t.Errorf("p1 body = %q, want %q", got.RawBody, want)
		}

		// The column itself must no longer hold the plaintext.
		var stored []byte
		if err := repo.pool.QueryRow(ctx,
			`SELECT raw_body FROM webhook_payload WHERE id = 'p1'`).Scan(&stored); err != nil {
			t.Fatalf("read raw column: %v", err)
		}
		if string(stored) == "From: a@x\nSubject: Secret One\n\nbody one" {
			t.Error("raw_body still contains plaintext")
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		stats, err := repo.BackfillEncryption(ctx, 10, false, nil)
		if err != nil {
			t.Fatalf("second run: %v", err)
		}
		if stats.Encrypted != 0 {
			t.Errorf("second run encrypted %d rows, want 0", stats.Encrypted)
		}
	})

	t.Run("picks up newly added plaintext rows", func(t *testing.T) {
		insertPlaintext(t, repo, ctx, "p4", "Subject: Later Secret")
		stats, err := repo.BackfillEncryption(ctx, 10, false, nil)
		if err != nil {
			t.Fatalf("third run: %v", err)
		}
		if stats.Encrypted != 1 {
			t.Errorf("encrypted %d rows, want 1", stats.Encrypted)
		}
	})
}

func TestBackfillEncryption_RequiresEncryptor(t *testing.T) {
	repo := &WebhookPayloadRepo{}
	if _, err := repo.BackfillEncryption(context.Background(), 10, false, nil); err == nil {
		t.Error("expected an error when no encryptor is configured")
	}
}

func TestBackfillEncryption_PaginatesPastMultipleBatches(t *testing.T) {
	repo, ctx := newBackfillTestRepo(t)

	for _, id := range []string{"a1", "a2", "a3", "a4", "a5"} {
		insertPlaintext(t, repo, ctx, id, "body "+id)
	}

	// A batch size smaller than the row count exercises the keyset cursor.
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
