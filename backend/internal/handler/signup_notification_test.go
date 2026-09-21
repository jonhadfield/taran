package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/testutil"
)

type sentNotification struct{ to, user, via string }

func checkAccessWith(t *testing.T, invite *domain.Invite, open, firstAcceptance bool, sendErr error) (int, []sentNotification) {
	t.Helper()
	var sent []sentNotification
	settings := memSettings{auth.OpenRegistrationSetting: "false"}
	if open {
		settings[auth.OpenRegistrationSetting] = "true"
	}
	h := &InviteHandler{
		Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(context.Context, string) (*domain.Invite, error) { return invite, nil },
			MarkAcceptedFn: func(context.Context, string) (bool, error) {
				return firstAcceptance, nil
			},
		},
		Settings:    settings,
		AdminEmails: []string{"owner@example.com", "second@example.com"},
		Mailer: &testutil.MockMailer{
			SendSignupNotificationFn: func(_ context.Context, to, user, via string) error {
				sent = append(sent, sentNotification{to, user, via})
				return sendErr
			},
		},
	}
	req := httptest.NewRequest("GET", "/api/access", nil)
	req = req.WithContext(contextWithUserEmail("u", "New@Example.com"))
	rec := httptest.NewRecorder()
	h.CheckAccess(rec, req)
	return rec.Code, sent
}

func TestCheckAccess_NotifiesAdminsOnOpenRegistrationSignup(t *testing.T) {
	code, sent := checkAccessWith(t, nil, true, true, nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	want := []sentNotification{
		{"owner@example.com", "new@example.com", "open registration"},
		{"second@example.com", "new@example.com", "open registration"},
	}
	if len(sent) != 2 || sent[0] != want[0] || sent[1] != want[1] {
		t.Errorf("sent = %+v, want %+v", sent, want)
	}
}

func TestCheckAccess_NotifiesOnFirstInvitedSignIn(t *testing.T) {
	invite := &domain.Invite{Email: "new@example.com", InvitedBy: "admin-user"}
	_, sent := checkAccessWith(t, invite, false, true, nil)
	if len(sent) != 2 || sent[0].via != "an invite" {
		t.Errorf("sent = %+v, want an invite notification per admin", sent)
	}
}

func TestCheckAccess_NoNotificationUnlessFirstAcceptance(t *testing.T) {
	accepted := time.Now()
	tests := map[string]struct {
		invite *domain.Invite
		first  bool
	}{
		// Another request marked it accepted first; that request notifies.
		"lost the race":    {&domain.Invite{Email: "new@example.com"}, false},
		"already accepted": {&domain.Invite{Email: "new@example.com", AcceptedAt: &accepted}, true},
	}
	for name, tt := range tests {
		if _, sent := checkAccessWith(t, tt.invite, false, tt.first, nil); len(sent) != 0 {
			t.Errorf("%s: sent %+v, want nothing", name, sent)
		}
	}
}

func TestCheckAccess_NotificationFailureDoesNotBlockAccess(t *testing.T) {
	code, sent := checkAccessWith(t, nil, true, true, errors.New("resend down"))
	if code != http.StatusOK {
		t.Errorf("status = %d, want 200 despite mail failure", code)
	}
	if len(sent) != 2 {
		t.Errorf("a failure for one admin should not stop the others: sent %+v", sent)
	}
}

func TestCheckAccess_DeniedUserTriggersNothing(t *testing.T) {
	if _, sent := checkAccessWith(t, nil, false, true, nil); len(sent) != 0 {
		t.Errorf("sent %+v for a user who was turned away", sent)
	}
}
