package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// Encryptor provides AES-256-GCM encryption and decryption using a master key.
type Encryptor struct {
	aead cipher.AEAD
}

// NewEncryptor creates an Encryptor from a hex-encoded 32-byte master key.
func NewEncryptor(masterKeyHex string) (*Encryptor, error) {
	key, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &Encryptor{aead: aead}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random 12-byte nonce.
func (e *Encryptor) Encrypt(plaintext string) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext = e.aead.Seal(nil, nonce, []byte(plaintext), nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided nonce.
func (e *Encryptor) Decrypt(ciphertext, nonce []byte) (string, error) {
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// EncryptToString encrypts plaintext and returns a base64-encoded string
// containing the nonce prepended to the ciphertext: base64(nonce + ciphertext).
func (e *Encryptor) EncryptToString(plaintext string) (string, error) {
	ciphertext, nonce, err := e.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	combined := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptFromString decodes a base64-encoded string containing nonce + ciphertext
// and returns the decrypted plaintext.
func (e *Encryptor) DecryptFromString(encoded string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	nonceSize := e.aead.NonceSize()
	if len(combined) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]
	return e.Decrypt(ciphertext, nonce)
}
