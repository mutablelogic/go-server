package types

import (
	"strings"
	"unicode"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// IsSingleQuoted checks if a string is single-quoted
func IsSingleQuoted(s string) bool {
	return s != "'" && strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")
}

// IsDoubleQuoted checks if a string is double-quoted
func IsDoubleQuoted(s string) bool {
	return s != "\"" && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")
}

// Quote returns a string with the input string quoted and escaped for
func Quote(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}

// DoubleQuote returns a string with the input string quoted and escaped for
func DoubleQuote(str string) string {
	return "\"" + strings.ReplaceAll(str, "\"", "\"\"") + "\""
}

// IsNumeric checks if a string consists only of digits
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsUppercase checks if a string is all in uppercase (A-Z)
func IsUppercase(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
