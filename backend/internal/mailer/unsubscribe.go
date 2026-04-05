package mailer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateUnsubscribeURL creates an HMAC-signed URL for one-click unsubscribe.
// The token includes the current month so it expires after ~2 months.
func GenerateUnsubscribeURL(baseURL, userID, secret string) string {
	period := currentPeriod()
	token := generateToken(userID, period, secret)
	return fmt.Sprintf("%s/api/public/unsubscribe?uid=%s&token=%s", baseURL, userID, token)
}

// ValidateUnsubscribeToken checks that the token matches the expected HMAC
// for the user. Accepts tokens from the current or previous month to handle
// emails in transit.
func ValidateUnsubscribeToken(userID, token, secret string) bool {
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
