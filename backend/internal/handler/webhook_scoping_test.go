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

// Regression: GetByMessageID and UpdateThreadID were not scoped by user.
//
// Message-IDs are not globally unique — a broadcast newsletter reaches every
// recipient carrying the same one — so a global dedupe check dropped the second
// and later recipients' copies as "duplicate", and the threading walk let
// attacker-controlled In-Reply-To/References headers on inbound mail reach and
// backfill the thread_id of another user's email.

func TestIngestEmail_DuplicateCheckIsUserScoped(t *testing.T) {
	var lookupUserID string
	var created bool

	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			// The same Message-ID already exists, but for a different user, so a
			// correctly scoped lookup must not find it.
			GetByMessageIDFn: func(_ context.Context, userID, _ string) (*domain.Email, error) {
				lookupUserID = userID
				if userID == "user-1" {
					return nil, fmt.Errorf("not found")
				}
				return &domain.Email{ID: "other-users-copy", UserID: "user-2"}, nil
			},
			CreateFn: func(_ context.Context, _ *domain.Email) error {
				created = true
				return nil
			},
		},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:      "newsletter@example.com",
		To:        "inbox@test.com",
		Subject:   "Weekly roundup",
		MessageID: "<broadcast-1@example.com>",
		TextBody:  "Body",
	})

	rec := httptest.NewRecorder()
	h.IngestEmail(rec, httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body)))

	if lookupUserID != "user-1" {
		t.Errorf("dedupe lookup used userID %q, want %q", lookupUserID, "user-1")
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d — a Message-ID owned by another user must not mask this delivery",
			rec.Code, http.StatusAccepted)
	}
	if !created {
		t.Error("email was not created; it was wrongly treated as a duplicate")
	}
}

func TestIngestEmail_ThreadBackfillIsUserScoped(t *testing.T) {
	var threadUpdateUserID string
	var lookupUserIDs []string

	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, userID, messageID string) (*domain.Email, error) {
				lookupUserIDs = append(lookupUserIDs, userID)
				// The referenced parent exists in this user's own mailbox and has
				// no thread_id yet, which triggers the backfill path.
				if messageID == "<parent@example.com>" {
					return &domain.Email{ID: "parent-1", UserID: "user-1", MessageID: messageID}, nil
				}
				return nil, fmt.Errorf("not found")
			},
			UpdateThreadIDFn: func(_ context.Context, userID, _, _ string) error {
				threadUpdateUserID = userID
				return nil
			},
			CreateFn: func(_ context.Context, _ *domain.Email) error { return nil },
		},
	}

	raw := []byte("From: sender@example.com\r\n" +
		"To: inbox@test.com\r\n" +
		"Subject: Re: thread\r\n" +
		"Message-ID: <reply@example.com>\r\n" +
		"In-Reply-To: <parent@example.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\nReply body")

	rec := httptest.NewRecorder()
	h.IngestEmail(rec, httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(raw)))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	for _, id := range lookupUserIDs {
		if id != "user-1" {
			t.Errorf("threading lookup used userID %q, want %q", id, "user-1")
		}
	}
	if threadUpdateUserID != "user-1" {
		t.Errorf("thread_id backfill used userID %q, want %q — an unscoped update can mutate another user's email",
			threadUpdateUserID, "user-1")
	}
}
