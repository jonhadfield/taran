package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hadfielj/taran/backend/internal/testutil"
)

type sentFeedback struct{ to, from, message string }

func sendFeedback(t *testing.T, h *UserFeedbackHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/feedback", strings.NewReader(body))
	req = req.WithContext(contextWithUserEmail("u1", "reader@example.com"))
	rec := httptest.NewRecorder()
	h.Send(rec, req)
	return rec
}

func TestUserFeedback_EmailsEveryAdminWithSenderAndMessage(t *testing.T) {
	var sent []sentFeedback
	h := &UserFeedbackHandler{
		AdminEmails: []string{"owner@example.com", "second@example.com"},
		Mailer: &testutil.MockMailer{
			SendUserFeedbackFn: func(_ context.Context, to, from, message string) error {
				sent = append(sent, sentFeedback{to, from, message})
				return nil
			},
		},
	}

	rec := sendFeedback(t, h, `{"Message":"  The digests are great, more detail on finance please  "}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	want := "The digests are great, more detail on finance please"
	if len(sent) != 2 {
		t.Fatalf("sent %d messages, want one per admin", len(sent))
	}
	for i, s := range sent {
		if s.message != want {
			t.Errorf("message %d = %q, want it trimmed to %q", i, s.message, want)
		}
		if s.from != "reader@example.com" {
			t.Errorf("from %d = %q, want the signed-in user", i, s.from)
		}
	}
	if sent[0].to != "owner@example.com" || sent[1].to != "second@example.com" {
		t.Errorf("recipients = %v", sent)
	}
}

func TestUserFeedback_Validation(t *testing.T) {
	for name, body := range map[string]string{
		"empty":      `{"Message":""}`,
		"whitespace": `{"Message":"   "}`,
		"too long":   `{"Message":"` + strings.Repeat("a", maxFeedbackLength+1) + `"}`,
		"bad json":   `{`,
	} {
		h := &UserFeedbackHandler{
			AdminEmails: []string{"owner@example.com"},
			Mailer: &testutil.MockMailer{
				SendUserFeedbackFn: func(context.Context, string, string, string) error {
					t.Errorf("%s: nothing should be sent", name)
					return nil
				},
			},
		}
		if rec := sendFeedback(t, h, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
}

func TestUserFeedback_UnavailableWithoutMailerOrAdmins(t *testing.T) {
	cases := map[string]*UserFeedbackHandler{
		"no mailer": {AdminEmails: []string{"owner@example.com"}},
		"no admins": {Mailer: &testutil.MockMailer{}},
	}
	for name, h := range cases {
		if rec := sendFeedback(t, h, `{"Message":"hello"}`); rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: status = %d, want 503", name, rec.Code)
		}
	}
}

func TestUserFeedback_OneFailedRecipientStillDelivers(t *testing.T) {
	h := &UserFeedbackHandler{
		AdminEmails: []string{"broken@example.com", "owner@example.com"},
		Mailer: &testutil.MockMailer{
			SendUserFeedbackFn: func(_ context.Context, to, _, _ string) error {
				if to == "broken@example.com" {
					return errors.New("mailbox full")
				}
				return nil
			},
		},
	}
	if rec := sendFeedback(t, h, `{"Message":"hello"}`); rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 when another admin received it", rec.Code)
	}
}

func TestUserFeedback_AllRecipientsFailing(t *testing.T) {
	h := &UserFeedbackHandler{
		AdminEmails: []string{"owner@example.com"},
		Mailer: &testutil.MockMailer{
			SendUserFeedbackFn: func(context.Context, string, string, string) error {
				return errors.New("provider down")
			},
		},
	}
	if rec := sendFeedback(t, h, `{"Message":"hello"}`); rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}
