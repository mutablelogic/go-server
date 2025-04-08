package schema

import "strings"

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// https://github.com/prometheus/OpenMetrics/blob/main/specification/OpenMetrics.md
	ContentTypeMetrics = "application/openmetrics-text; version=1.0.0; charset=utf-8"
)

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func escapeString(s string) string {
	// Escape special characters
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// Contains checks if a string contains a set of runes
func stringContains(s string, fn func(i int, r rune) bool) bool {
	for i, r := range s {
		if !fn(i, r) {
			return false
		}
	}
	return true
}
