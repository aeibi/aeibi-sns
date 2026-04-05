package util

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/rrivera/identicon"
)

const (
	defaultAvatarNamespace = "aeibi"
	defaultAvatarSize      = 5
	defaultAvatarDensity   = 3
	defaultAvatarPixels    = 256
)

// GenerateDefaultAvatar returns a PNG-encoded identicon derived from uid.
func GenerateDefaultAvatar(uid string) ([]byte, error) {
	if uid == "" {
		return nil, errors.New("uid is empty")
	}

	generator, err := identicon.New(defaultAvatarNamespace, defaultAvatarSize, defaultAvatarDensity)
	if err != nil {
		return nil, fmt.Errorf("init identicon generator: %w", err)
	}

	icon, err := generator.Draw(uid)
	if err != nil {
		return nil, fmt.Errorf("draw identicon: %w", err)
	}

	var buf bytes.Buffer
	if err := icon.Png(defaultAvatarPixels, &buf); err != nil {
		return nil, fmt.Errorf("encode identicon: %w", err)
	}

	return buf.Bytes(), nil
}
