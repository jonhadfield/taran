package mailer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"time"
)

// GenerateUnsubscribeURL creates an HMAC-signed URL for one-click unsubscribe.
// The token includes the current month so it expires after ~2 months.
// Returns "" if no secret is configured, since an unsigned link would be
// forgeable by anyone who knows a user ID.
func GenerateUnsubscribeURL(baseURL, userID, secret string) string {
	if secret == "" {
		return ""
	}
	period := currentPeriod()
	token := generateToken(userID, period, secret)
	return fmt.Sprintf("%s/api/public/unsubscribe?uid=%s&token=%s",
		baseURL, url.QueryEscape(userID), url.QueryEscape(token))
}

// ValidateUnsubscribeToken checks that the token matches the expected HMAC
// for the user. Accepts tokens from the current or previous month to handle
// emails in transit.
//
// Fails closed when no secret is configured: HMAC with an empty key is still a
// well-defined value an attacker can compute, so accepting it would let anyone
// holding a user ID unsubscribe that user.
func ValidateUnsubscribeToken(userID, token, secret string) bool {
	if secret == "" || token == "" {
		return false
	}
	current := currentPeriod()
	if hmac.Equal([]byte(generateToken(userID, current, secret)), []byte(token)) {
		return true
	}
	// Also accept last month's token for emails in transit
	prev := previousPeriod()
	return hmac.Equal([]byte(generateToken(userID, prev, secret)), []byte(token))
}

func generateToken(userID, period, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID + ":" + period))
	return hex.EncodeToString(mac.Sum(nil))
}

func currentPeriod() string {
	return time.Now().UTC().Format("2006-01")
}

func previousPeriod() string {
	return time.Now().UTC().AddDate(0, -1, 0).Format("2006-01")
}
