package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// BoolToInt64 converts a boolean to 1 or 0.
func BoolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

// SHA256 returns the hex-encoded SHA-256 hash of the given data.
func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
