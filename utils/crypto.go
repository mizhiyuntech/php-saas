package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func DecryptAES256GCM(apiKey, nonce, ciphertext, associatedData string) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}

	block, err := aes.NewCipher([]byte(apiKey))
	if err != nil {
		return "", fmt.Errorf("aes cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm failed: %w", err)
	}

	plaintext, err := gcm.Open(nil, []byte(nonce), ciphertextBytes, []byte(associatedData))
	if err != nil {
		return "", fmt.Errorf("decrypt failed: %w", err)
	}

	return string(plaintext), nil
}
