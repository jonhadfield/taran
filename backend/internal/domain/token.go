package domain

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateShareToken produces a cryptographically random 32-character hex token
// suitable for use as a share token for digests.
func GenerateShareToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
