package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

func AES256Encrypt(plaintext, key string) (string, error) {
	keyBytes := deriveKey(key)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func AES256Decrypt(cipherB64, key string) (string, error) {
	keyBytes := deriveKey(key)
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

func deriveKey(key string) []byte {
	b := []byte(key)
	if len(b) >= 32 {
		return b[:32]
	}
	result := make([]byte, 32)
	copy(result, b)
	return result
}

func EncryptJSON(data interface{}, key string) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return AES256Encrypt(string(jsonBytes), key)
}

func DecryptJSON(cipherB64, key string, target interface{}) error {
	plaintext, err := AES256Decrypt(cipherB64, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(plaintext), target)
}
