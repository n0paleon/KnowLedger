package utils

import (
	"strings"
	"testing"
)

func TestSHA256Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "known string",
			input:    []byte("hello"),
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0xFF, 0xAB},
			expected: "698ec68f9728531a9a2fd81d0c3cfe71b125d299f21442b95adf3c1a843f605a",
		},
		{
			name:     "longer text",
			input:    []byte("Hello, CAS with Cloudflare R2!"),
			expected: "9e1848144d8dea304dcef4580dbc844930b670cba57976e696d5cc86a18f09c5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SHA256Hash(tt.input)
			if got != tt.expected {
				t.Errorf("SHA256Hash(%q)\ngot:  %s\nwant: %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSHA256Hash_OutputFormat(t *testing.T) {
	hash := SHA256Hash([]byte("test"))

	if len(hash) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(hash))
	}
	if hash != strings.ToLower(hash) {
		t.Errorf("expected lowercase hex, got %s", hash)
	}
}

func TestSHA256Hash_Deterministic(t *testing.T) {
	input := []byte("consistent input")

	hash1 := SHA256Hash(input)
	hash2 := SHA256Hash(input)

	if hash1 != hash2 {
		t.Errorf("non-deterministic: got %s then %s", hash1, hash2)
	}
}

func TestSHA256Hash_Uniqueness(t *testing.T) {
	h1 := SHA256Hash([]byte("hello"))
	h2 := SHA256Hash([]byte("hello "))
	h3 := SHA256Hash([]byte("Hello"))

	if h1 == h2 || h1 == h3 || h2 == h3 {
		t.Error("different inputs produced the same hash (collision detected)")
	}
}
