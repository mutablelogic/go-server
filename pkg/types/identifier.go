package types

import "regexp"

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// An identifier is a string which starts with a letter and is followed by
	// letters, numbers, underscores or hyphens. It must be between 1 and 32 characters
	ReIdentifier = `[a-zA-Z][a-zA-Z0-9_\-]{0,31}`
)

var (
	reValidIdentifier = regexp.MustCompile(`^` + ReIdentifier + `$`)
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the string is a valid identifier
func IsIdentifier(s string) bool {
	return reValidIdentifier.MatchString(s)
}
