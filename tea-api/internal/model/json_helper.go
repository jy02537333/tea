package model

import (
	"encoding/json"
	"strings"
)

// NormalizeJSONOrNull ensures the returned string is valid JSON text or the literal null.
// Empty strings or whitespace are coerced to "null" to keep MySQL JSON columns happy.
func NormalizeJSONOrNull(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "null"
	}
	if trimmed == "null" {
		return "null"
	}
	if json.Valid([]byte(trimmed)) {
		return trimmed
	}
	return "null"
}
