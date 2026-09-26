package database

import (
	"context"
	"fmt"
	"strings"
)

// UserEmailLookup resolves a user's email address. Background paths have no
// request context, so they cannot read the email from a session.
type UserEmailLookup interface {
	GetUserEmail(ctx context.Context, userID string) (string, error)
}

// AdminOnlyAnalysisRules restricts analysis rules to admin accounts.
//
// The API routes are already behind AdminOnly, so no one else can create a
// rule. This closes the other half: rules created before the feature was
// restricted stay in the database but are never sent to the model. Gating here
// rather than in the caller means every path that applies rules — per-email
// extraction and digest generation alike — is covered by one check.
type AdminOnlyAnalysisRules struct {
	Rules       ActiveAnalysisRules
	Users       UserEmailLookup
	AdminEmails []string
}

// ListActiveRules returns the user's active rules when they are an admin, and
// no rules otherwise. A nil receiver or missing collaborator disables rules,
// matching the "nil disables them" convention the callers already use.
func (g *AdminOnlyAnalysisRules) ListActiveRules(ctx context.Context, userID string) ([]string, error) {
	if g == nil || g.Rules == nil || g.Users == nil || len(g.AdminEmails) == 0 {
		return nil, nil
	}

	email, err := g.Users.GetUserEmail(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("analysis rules admin check: %w", err)
	}

	email = strings.ToLower(strings.TrimSpace(email))
	for _, admin := range g.AdminEmails {
		if email != "" && email == strings.ToLower(strings.TrimSpace(admin)) {
			return g.Rules.ListActiveRules(ctx, userID)
		}
	}
	return nil, nil
}
