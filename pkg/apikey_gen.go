package pkg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}
	return "sk-" + hex.EncodeToString(bytes), nil
}

func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:6] + "..." + key[len(key)-4:]
}
