package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"

	"golang.org/x/crypto/nacl/secretbox"
)

// EncID encrypts uint32 id into an opaque, URL-safe token.
// Returns "" on failure.
func EncID(id uint32, secret string) string {
	// Derive 32-byte key from secret (lightweight KDF).
	sum := sha256.Sum256([]byte(secret))
	var key [32]byte
	copy(key[:], sum[:])

	// Plaintext: 4 bytes.
	var msg [4]byte
	binary.BigEndian.PutUint32(msg[:], id)

	// Nonce: 24 bytes (must be unique per key; random is safe & easy).
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return ""
	}

	// Token bytes: nonce || secretbox(msg).
	out := secretbox.Seal(nonce[:], msg[:], &nonce, &key)

	return base64.RawURLEncoding.EncodeToString(out)
}

// DecID decrypts token back to uint32.
// Returns (0,false) if invalid/tampered/wrong secret.
func DecID(tok string, secret string) (uint32, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil || len(raw) < 24+secretbox.Overhead {
		return 0, false
	}

	sum := sha256.Sum256([]byte(secret))
	var key [32]byte
	copy(key[:], sum[:])

	var nonce [24]byte
	copy(nonce[:], raw[:24])
	box := raw[24:]

	opened, ok := secretbox.Open(nil, box, &nonce, &key)
	if !ok || len(opened) != 4 {
		return 0, false
	}

	return binary.BigEndian.Uint32(opened), true
}
