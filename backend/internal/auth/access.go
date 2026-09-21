package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

const (
	// OpenRegistrationSetting is the app_setting key that, when true, lets any
	// signed-in user in without an invite.
	OpenRegistrationSetting = "open_registration"
	// OpenRegistrationInviter is recorded as invited_by on invites created for
	// users who got in through open registration. It lets those users keep
	// access once open registration is turned off, and lets admins count them.
	OpenRegistrationInviter = "open-registration"
)

// Reasons returned by AccessChecker.Check.
const (
	AccessAdmin            = "admin"
	AccessInvited          = "invited"
	AccessOpenRegistration = "open_registration"
	AccessNotInvited       = "not_invited"
)

// BoolSettings reads boolean app settings.
type BoolSettings interface {
	GetBool(ctx context.Context, key string, fallback bool) (bool, error)
}

// Access is the outcome of an access check.
type Access struct {
	Allowed bool
	Reason  string         // one of the Access* reasons
	Invite  *domain.Invite // the user's invite, if they have one
}

// AccessChecker decides whether a signed-in user may use the app: admins and
// invited users always may, and anyone else may while open registration is on.
type AccessChecker struct {
	Invites     database.InviteRepository
	Settings    BoolSettings // nil means open registration is never on
	AdminEmails []string
}

// Check reports whether email has access and why. A user admitted through open
// registration is given an invite, so later checks see them as invited.
func (c *AccessChecker) Check(ctx context.Context, email string) (Access, error) {
	email = strings.ToLower(email)
	for _, admin := range c.AdminEmails {
		if email == admin {
			return Access{Allowed: true, Reason: AccessAdmin}, nil
		}
	}

	invite, err := c.Invites.GetByEmail(ctx, email)
	if err != nil {
		return Access{}, err
	}
	if invite != nil {
		return Access{Allowed: true, Reason: AccessInvited, Invite: invite}, nil
	}

	denied := Access{Reason: AccessNotInvited}
	if c.Settings == nil {
		return denied, nil
	}
	open, err := c.Settings.GetBool(ctx, OpenRegistrationSetting, false)
	if err != nil {
		return Access{}, err
	}
	if !open {
		return denied, nil
	}

	invite = &domain.Invite{
		ID:        uuid.New().String(),
		Email:     email,
		InvitedBy: OpenRegistrationInviter,
		CreatedAt: time.Now().UTC(),
	}
	if err := c.Invites.Create(ctx, invite); err != nil {
		// A concurrent request for the same new user may have created the
		// invite first (email is unique); that still means they're in.
		existing, getErr := c.Invites.GetByEmail(ctx, email)
		if getErr != nil || existing == nil {
			return Access{}, fmt.Errorf("record open-registration invite: %w", err)
		}
		invite = existing
	}
	return Access{Allowed: true, Reason: AccessOpenRegistration, Invite: invite}, nil
}
