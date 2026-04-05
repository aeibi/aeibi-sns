package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

const defaultRandomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString generates a cryptographically secure random string of the given length using defaultRandomCharset.
func RandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	charsetLength := big.NewInt(int64(len(defaultRandomCharset)))
	b := make([]byte, length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("generate random index: %w", err)
		}

		b[i] = defaultRandomCharset[idx.Int64()]
	}

	return string(b), nil
}

func RandomString64() (string, error) {
	charsetLength := big.NewInt(int64(len(defaultRandomCharset)))
	b := make([]byte, 64)

	for i := 0; i < 64; i++ {
		idx, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("generate random index: %w", err)
		}

		b[i] = defaultRandomCharset[idx.Int64()]
	}

	return string(b), nil
}
