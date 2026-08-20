package handler

import "testing"

// Regression: any 3-30 char name was claimable, including the RFC 2142 role
// addresses and the admin/webmaster/hostmaster/postmaster addresses that
// certificate authorities use for domain validation.
func TestValidateUsername_RejectsReserved(t *testing.T) {
	reserved := []string{
		"admin", "administrator", "webmaster", "hostmaster", "postmaster",
		"abuse", "security", "noreply", "mailer-daemon", "digest", "support",
	}
	for _, name := range reserved {
		if err := validateUsername(name); err == nil {
			t.Errorf("validateUsername(%q) = nil, want an error", name)
		}
	}
}

func TestValidateUsername_AllowsOrdinaryNames(t *testing.T) {
	for _, name := range []string{"jon", "jon-hadfield", "user123", "abc"} {
		if err := validateUsername(name); err != nil {
			t.Errorf("validateUsername(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateUsername_EnforcesShape(t *testing.T) {
	for _, name := range []string{"ab", "-leading", "trailing-", "Upper", "has_underscore", ""} {
		if err := validateUsername(name); err == nil {
			t.Errorf("validateUsername(%q) = nil, want an error", name)
		}
	}
}
