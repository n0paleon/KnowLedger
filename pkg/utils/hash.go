package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// SHA256Hash returns a unique hash based on the given data.
// NOTE: the same data will always produce the same hash.
func SHA256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// KeyFromHash converts a hash string into a sharded object key
// using the first 2 characters as prefix for efficient storage listing.
//
// Format: "{2-char-prefix}/{full-hash}"
//
// Example:
//
//	KeyFromHash("2cf24dba...") == "2c/2cf24dba..."
//
// Panics if hash length < 2.
func KeyFromHash(hash string) string {
	return hash[:2] + "/" + hash
}

// HashContent computes the SHA-256 hash of data and returns
// the sharded object key in format "{2-char-prefix}/{hash}",
// ready to use as a Key in PutObject / GetObject.
//
// Example:
//
//	key := HashContent([]byte("hello"))
//	// key == "2c/2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
func HashContent(data []byte) string {
	return KeyFromHash(SHA256Hash(data))
}

// GeneratePasswordHash is simplified version for bcrypt.GenerateFromPassword
func GeneratePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash is simplified version of bcrypt.CompareHashAndPassword
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
