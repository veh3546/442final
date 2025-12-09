package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// genRandomHex returns a hex-encoded random byte string of length 2*n.
func genRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// genRandomBase64URL returns a URL-safe base64 (no padding) encoded random byte string.
// For n bytes, the output length is ceil(4*n/3) (e.g., 32 bytes -> 44 chars), which fits in VARCHAR(50).
func genRandomBase64URL(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
