package types

import "regexp"

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	ReIdentifier = `[a-zA-Z][a-zA-Z0-9_\-]*`
)

var (
	reValidName = regexp.MustCompile(`^` + ReIdentifier + `$`)
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the string is a valid identifier
func IsIdentifier(s string) bool {
	return reValidName.MatchString(s)
}
