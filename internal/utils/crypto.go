package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

// GetEncryptionKey retrieves the encryption key from environment variable or returns a default fallback key.
// AES-256 requires a 32-byte key.
func GetEncryptionKey() []byte {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if len(keyStr) >= 32 {
		return []byte(keyStr[:32])
	}
	// Fallback 32-byte key for local development
	return []byte("abcdefghijklmnopqrstuvwxyz012345")
}

// Encrypt encrypts a plaintext string using AES-GCM and returns a hex-encoded ciphertext.
func Encrypt(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(cipherText), nil
}

// Decrypt decrypts a hex-encoded ciphertext using AES-GCM and returns the plaintext.
func Decrypt(cipherTextHex string, key []byte) (string, error) {
	cipherText, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
