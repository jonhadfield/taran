package auth_test

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

// fakeSettings serves the open_registration flag.
type fakeSettings struct {
	open bool
	err  error
}

func (f fakeSettings) GetBool(_ context.Context, key string, fallback bool) (bool, error) {
	if key != auth.OpenRegistrationSetting {
		return fallback, nil
	}
	return f.open, f.err
}

// inviteStore is an in-memory invite repo keyed by email.
func inviteStore() (*testutil.MockInviteRepo, map[string]*domain.Invite) {
	invites := map[string]*domain.Invite{}
	return &testutil.MockInviteRepo{
		GetByEmailFn: func(_ context.Context, email string) (*domain.Invite, error) {
			return invites[email], nil
		},
		CreateFn: func(_ context.Context, inv *domain.Invite) error {
			if _, ok := invites[inv.Email]; ok {
				return errors.New("duplicate key value violates unique constraint")
			}
			invites[inv.Email] = inv
			return nil
		},
	}, invites
}

func TestAccessChecker_AdminAndInvited(t *testing.T) {
	repo, invites := inviteStore()
	invites["invited@example.com"] = &domain.Invite{Email: "invited@example.com", InvitedBy: "admin"}
	c := auth.AccessChecker{Invites: repo, Settings: fakeSettings{}, AdminEmails: []string{"admin@example.com"}}

	for email, want := range map[string]string{
		"Admin@Example.com":    auth.AccessAdmin,
		"invited@example.com":  auth.AccessInvited,
		"stranger@example.com": auth.AccessNotInvited,
	} {
		got, err := c.Check(context.Background(), email)
		if err != nil {
			t.Fatalf("%s: %v", email, err)
		}
		if got.Reason != want || got.Allowed != (want != auth.AccessNotInvited) {
			t.Errorf("%s: got %+v, want reason %q", email, got, want)
		}
	}
	if len(invites) != 1 {
		t.Errorf("invites were created while open registration was off: %v", invites)
	}
}

func TestAccessChecker_OpenRegistrationCreatesTaggedInvite(t *testing.T) {
	repo, invites := inviteStore()
	c := auth.AccessChecker{Invites: repo, Settings: fakeSettings{open: true}}

	got, err := c.Check(context.Background(), "New.User@Example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Allowed || got.Reason != auth.AccessOpenRegistration {
		t.Fatalf("got %+v, want allowed via open registration", got)
	}
	inv := invites["new.user@example.com"]
	if inv == nil || inv.InvitedBy != auth.OpenRegistrationInviter || inv.ID == "" || inv.CreatedAt.IsZero() {
		t.Fatalf("invite = %+v, want a tagged invite for the lowercased email", inv)
	}
	if got.Invite != inv {
		t.Error("Check should return the invite it created")
	}

	// Once recorded, the user is simply invited, so they keep access after the
	// trial ends.
	c.Settings = fakeSettings{open: false}
	again, err := c.Check(context.Background(), "new.user@example.com")
	if err != nil || !again.Allowed || again.Reason != auth.AccessInvited {
		t.Errorf("after closing: got %+v, %v; want still invited", again, err)
	}
}

func TestAccessChecker_ConcurrentCreateStillAllowed(t *testing.T) {
	existing := &domain.Invite{Email: "racer@example.com", InvitedBy: auth.OpenRegistrationInviter}
	lookups := 0
	repo := &testutil.MockInviteRepo{
		GetByEmailFn: func(_ context.Context, _ string) (*domain.Invite, error) {
			lookups++
			if lookups == 1 {
				return nil, nil // not there yet
			}
			return existing, nil // another request created it meanwhile
		},
		CreateFn: func(context.Context, *domain.Invite) error {
			return errors.New("duplicate key value violates unique constraint")
		},
	}
	c := auth.AccessChecker{Invites: repo, Settings: fakeSettings{open: true}}

	got, err := c.Check(context.Background(), "racer@example.com")
	if err != nil || !got.Allowed || got.Invite != existing {
		t.Errorf("got %+v, %v; want allowed with the existing invite", got, err)
	}
}

func TestAccessChecker_Errors(t *testing.T) {
	failingCreate := &testutil.MockInviteRepo{
		CreateFn: func(context.Context, *domain.Invite) error { return errors.New("db down") },
	}
	tests := map[string]auth.AccessChecker{
		"settings error": {Invites: &testutil.MockInviteRepo{}, Settings: fakeSettings{err: errors.New("db down")}},
		"create error":   {Invites: failingCreate, Settings: fakeSettings{open: true}},
		"lookup error": {Invites: &testutil.MockInviteRepo{
			GetByEmailFn: func(context.Context, string) (*domain.Invite, error) { return nil, errors.New("db down") },
		}, Settings: fakeSettings{open: true}},
	}
	for name, c := range tests {
		if got, err := c.Check(context.Background(), "someone@example.com"); err == nil || got.Allowed {
			t.Errorf("%s: got %+v, %v; want error and no access", name, got, err)
		}
	}
}

func TestAccessChecker_NilSettingsIsInviteOnly(t *testing.T) {
	c := auth.AccessChecker{Invites: &testutil.MockInviteRepo{}}
	got, err := c.Check(context.Background(), "someone@example.com")
	if err != nil || got.Allowed {
		t.Errorf("got %+v, %v; want denied", got, err)
	}
}

func TestSessionAuth_OpenRegistrationGate(t *testing.T) {
	for _, tt := range []struct {
		open bool
		want int
	}{{false, http.StatusForbidden}, {true, http.StatusOK}} {
		sessions := &testutil.MockSessionRepo{
			GetByTokenFn: func(_ context.Context, token string) (*domain.Session, error) {
				return &domain.Session{ID: "s", UserID: "u", Token: token, UserEmail: "new@example.com",
					ExpiresAt: time.Now().Add(time.Hour), UpdatedAt: time.Now()}, nil
			},
		}
		repo, _ := inviteStore()
		sa := &auth.SessionAuth{Sessions: sessions, Invites: repo, Settings: fakeSettings{open: tt.open}}

		req := httptest.NewRequest("GET", "/api/emails", nil)
		req.Header.Set("Authorization", "Bearer t")
		rec := httptest.NewRecorder()
		sa.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(rec, req)

		if rec.Code != tt.want {
			t.Errorf("open=%v: status = %d, want %d", tt.open, rec.Code, tt.want)
		}
	}
}
