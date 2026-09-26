package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

// buildRawEmail assembles a message with a References chain of the given length,
// where the ancestor that exists sits at the given index.
func buildRawEmail(references string) []byte {
	return []byte("From: sender@example.com\r\n" +
		"To: inbox@test.com\r\n" +
		"Subject: Re: deep thread\r\n" +
		"Message-ID: <reply@example.com>\r\n" +
		"References: " + references + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\nReply body")
}

func TestIngestEmail_ReferencesResolvedInOneLookup(t *testing.T) {
	// A long References chain, as an inbound message may carry. Before this was
	// batched, each entry cost its own query, so the sender decided how many
	// round trips the ingest path made.
	var refs string
	for i := 1; i <= 50; i++ {
		refs += fmt.Sprintf("<ref%d@example.com> ", i)
	}

	lookups := 0
	var sawUserIDs []string
	var created *domain.Email

	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			FindThreadRefsFn: func(_ context.Context, userID string, ids []string) (map[string]domain.ThreadRef, error) {
				lookups++
				sawUserIDs = append(sawUserIDs, userID)
				// Only one of the fifty references exists, and it already
				// belongs to a thread.
				return map[string]domain.ThreadRef{
					"<ref7@example.com>": {ID: "e7", MessageID: "<ref7@example.com>", ThreadID: "<thread-root>"},
				}, nil
			},
			CreateFn: func(_ context.Context, e *domain.Email) error { created = e; return nil },
		},
	}

	rec := httptest.NewRecorder()
	h.IngestEmail(rec, httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(buildRawEmail(refs))))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if lookups != 1 {
		t.Errorf("reference lookups = %d, want 1 regardless of chain length", lookups)
	}
	if created == nil {
		t.Fatal("email was not created")
	}
	if created.ThreadID != "<thread-root>" {
		t.Errorf("ThreadID = %q, want the existing ancestor's thread", created.ThreadID)
	}
	// The threading headers are attacker-supplied, so the lookup must stay
	// scoped to the receiving user.
	for _, id := range sawUserIDs {
		if id != "user-1" {
			t.Errorf("lookup used userID %q, want %q", id, "user-1")
		}
	}
}

func TestIngestEmail_ReferencesPrefersMostRecentAncestor(t *testing.T) {
	// Two ancestors exist. References is ordered oldest first, so the later
	// entry must win — the same order the per-entry loop walked.
	var created *domain.Email
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			FindThreadRefsFn: func(_ context.Context, _ string, _ []string) (map[string]domain.ThreadRef, error) {
				return map[string]domain.ThreadRef{
					"<old@example.com>":    {ID: "e1", MessageID: "<old@example.com>", ThreadID: "<older-thread>"},
					"<recent@example.com>": {ID: "e2", MessageID: "<recent@example.com>", ThreadID: "<recent-thread>"},
				}, nil
			},
			CreateFn: func(_ context.Context, e *domain.Email) error { created = e; return nil },
		},
	}

	raw := buildRawEmail("<old@example.com> <recent@example.com>")
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(raw)))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if created == nil {
		t.Fatal("email was not created")
	}
	if created.ThreadID != "<recent-thread>" {
		t.Errorf("ThreadID = %q, want %q — References is ordered oldest first, so the last entry wins",
			created.ThreadID, "<recent-thread>")
	}
}

func TestIngestEmail_ReferencesLookupFailureStillAccepts(t *testing.T) {
	// A failed lookup must not drop the email; it just starts its own thread.
	var created *domain.Email
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			FindThreadRefsFn: func(_ context.Context, _ string, _ []string) (map[string]domain.ThreadRef, error) {
				return nil, fmt.Errorf("database unavailable")
			},
			CreateFn: func(_ context.Context, e *domain.Email) error { created = e; return nil },
		},
	}

	rec := httptest.NewRecorder()
	h.IngestEmail(rec, httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(buildRawEmail("<ref1@example.com>"))))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if created == nil {
		t.Fatal("email was dropped when the lookup failed")
	}
	if created.ThreadID != "<reply@example.com>" {
		t.Errorf("ThreadID = %q, want the email's own message id", created.ThreadID)
	}
}
