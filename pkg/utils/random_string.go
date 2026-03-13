package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/oklog/ulid/v2"
)

func GenerateRandomULID() string {
	return ulid.Make().String()
}

// GenerateAPIKey generate random api key
func GenerateAPIKey(length int, prefix string) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b)), nil
}
