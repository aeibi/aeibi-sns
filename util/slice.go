package util

import (
	"encoding/json"
	"strings"
)

// NormalizeStrings trims whitespace, deduplicates, and preserves order.
func NormalizeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		normalized = append(normalized, v)
	}
	return normalized
}

// DecodeStringSlice parses a JSON-encoded string slice stored as text.
func DecodeStringSlice(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	if values == nil {
		return []string{}, nil
	}
	return values, nil
}

// EncodeStringSlice serializes a string slice to JSON text, returning "[]" for nil/empty.
func EncodeStringSlice(values []string) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
