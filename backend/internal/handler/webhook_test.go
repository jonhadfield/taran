package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

func TestIngestEmail_Success(t *testing.T) {
	var createdEmail *domain.Email

	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
			CreateFn: func(_ context.Context, e *domain.Email) error {
				createdEmail = e
				return nil
			},
		},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:      "sender@example.com",
		To:        "inbox@test.com",
		Subject:   "Hello",
		MessageID: "<msg-1@example.com>",
		TextBody:  "Body text",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if createdEmail == nil {
		t.Fatal("email not created")
	}
	if createdEmail.Subject != "Hello" {
		t.Errorf("Subject = %q, want %q", createdEmail.Subject, "Hello")
	}
	if createdEmail.Status != domain.EmailStatusPending {
		t.Errorf("Status = %q, want %q", createdEmail.Status, domain.EmailStatusPending)
	}
}

func TestIngestEmail_InvalidMIME(t *testing.T) {
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{},
		Emails:   &testutil.MockEmailRepo{},
	}

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader([]byte("not valid mime")))
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestIngestEmail_UnknownRecipient(t *testing.T) {
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, _ string) (*domain.EmailAccount, error) {
				return nil, nil
			},
		},
		Emails: &testutil.MockEmailRepo{},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "sender@example.com",
		To:       "unknown@test.com",
		Subject:  "Hello",
		TextBody: "Body",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestIngestEmail_DuplicateMessageID(t *testing.T) {
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return &domain.Email{ID: "existing-id"}, nil // already exists
			},
		},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:      "sender@example.com",
		To:        "inbox@test.com",
		Subject:   "Hello",
		MessageID: "<duplicate@example.com>",
		TextBody:  "Body",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "duplicate" {
		t.Errorf("status = %q, want %q", resp["status"], "duplicate")
	}
}

func TestIngestEmail_CreateError(t *testing.T) {
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
			CreateFn: func(_ context.Context, _ *domain.Email) error {
				return fmt.Errorf("db error")
			},
		},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:      "sender@example.com",
		To:        "inbox@test.com",
		Subject:   "Hello",
		MessageID: "<msg-1@example.com>",
		TextBody:  "Body",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestIngestEmail_XOriginalToHeader(t *testing.T) {
	var createdEmail *domain.Email

	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
			CreateFn: func(_ context.Context, e *domain.Email) error {
				createdEmail = e
				return nil
			},
		},
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "sender@example.com",
		To:       "original@test.com",
		Subject:  "Hello",
		TextBody: "Body",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	req.Header.Set("X-Original-To", "override@test.com")
	rec := httptest.NewRecorder()
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if createdEmail == nil {
		t.Fatal("email not created")
	}
	if createdEmail.ToAddress != "override@test.com" {
		t.Errorf("ToAddress = %q, want %q", createdEmail.ToAddress, "override@test.com")
	}
}

func TestIngestEmail_NoProvider(t *testing.T) {
	h := &WebhookHandler{
		Accounts: &testutil.MockAccountRepo{
			GetByEmailAddressFn: func(_ context.Context, addr string) (*domain.EmailAccount, error) {
				return &domain.EmailAccount{ID: "acct-1", UserID: "user-1", EmailAddress: addr}, nil
			},
		},
		Emails: &testutil.MockEmailRepo{
			GetByMessageIDFn: func(_ context.Context, _, _ string) (*domain.Email, error) {
				return nil, fmt.Errorf("not found")
			},
		},
		// Provider is nil — extraction should be skipped gracefully
	}

	body := testutil.BuildRawEmail(testutil.RawEmailOpts{
		From:     "sender@example.com",
		To:       "inbox@test.com",
		Subject:  "Hello",
		TextBody: "Body",
	})

	req := httptest.NewRequest("POST", "/webhook/email", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	// Should not panic
	h.IngestEmail(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}
