package mailer

import "testing"

const testSecret = "unsubscribe-signing-secret"

func TestValidateUnsubscribeToken_AcceptsCurrentToken(t *testing.T) {
	token := generateToken("user-1", currentPeriod(), testSecret)
	if !ValidateUnsubscribeToken("user-1", token, testSecret) {
		t.Error("a freshly generated token should validate")
	}
}

func TestValidateUnsubscribeToken_AcceptsPreviousPeriod(t *testing.T) {
	token := generateToken("user-1", previousPeriod(), testSecret)
	if !ValidateUnsubscribeToken("user-1", token, testSecret) {
		t.Error("last month's token should still validate for mail in transit")
	}
}

func TestValidateUnsubscribeToken_RejectsOtherUsersToken(t *testing.T) {
	token := generateToken("user-2", currentPeriod(), testSecret)
	if ValidateUnsubscribeToken("user-1", token, testSecret) {
		t.Error("a token issued for another user must not validate")
	}
}

// Regression: with no secret configured the HMAC was computed with an empty
// key, which any attacker can reproduce — so knowing a user ID was enough to
// unsubscribe that user.
func TestValidateUnsubscribeToken_FailsClosedWithoutSecret(t *testing.T) {
	forged := generateToken("user-1", currentPeriod(), "")
	if ValidateUnsubscribeToken("user-1", forged, "") {
		t.Error("validation must fail closed when no secret is configured")
	}
}

func TestValidateUnsubscribeToken_RejectsEmptyToken(t *testing.T) {
	if ValidateUnsubscribeToken("user-1", "", testSecret) {
		t.Error("an empty token must not validate")
	}
}

func TestGenerateUnsubscribeURL_EmptyWithoutSecret(t *testing.T) {
	if got := GenerateUnsubscribeURL("https://example.com", "user-1", ""); got != "" {
		t.Errorf("URL = %q, want empty — an unsigned link would be forgeable", got)
	}
}

func TestGenerateUnsubscribeURL_RoundTrips(t *testing.T) {
	url := GenerateUnsubscribeURL("https://example.com", "user-1", testSecret)
	if url == "" {
		t.Fatal("expected a URL")
	}
	token := generateToken("user-1", currentPeriod(), testSecret)
	if !ValidateUnsubscribeToken("user-1", token, testSecret) {
		t.Error("generated link's token should validate")
	}
}
