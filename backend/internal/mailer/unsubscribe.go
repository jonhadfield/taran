package mailer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateUnsubscribeURL creates an HMAC-signed URL for one-click unsubscribe.
func GenerateUnsubscribeURL(baseURL, userID, secret string) string {
	token := generateToken(userID, secret)
	return fmt.Sprintf("%s/api/public/unsubscribe?uid=%s&token=%s", baseURL, userID, token)
}

// ValidateUnsubscribeToken checks that the token matches the expected HMAC for the user.
func ValidateUnsubscribeToken(userID, token, secret string) bool {
	expected := generateToken(userID, secret)
	return hmac.Equal([]byte(expected), []byte(token))
}

func generateToken(userID, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	return hex.EncodeToString(mac.Sum(nil))
}
