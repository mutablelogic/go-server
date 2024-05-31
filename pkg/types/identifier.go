package types

import "regexp"

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
<<<<<<< HEAD
	// An identifier is a string which starts with a letter and is followed by
	// letters, numbers, underscores or hyphens. It must be between 1 and 32 characters
	ReIdentifier = `[a-zA-Z][a-zA-Z0-9_\-]{0,31}`
)

var (
	reValidIdentifier = regexp.MustCompile(`^` + ReIdentifier + `$`)
=======
	ReIdentifier = `[a-zA-Z][a-zA-Z0-9_\-]*`
)

var (
	reValidName = regexp.MustCompile(`^` + ReIdentifier + `$`)
>>>>>>> a486469478ac5f8553b60a97d8eb2a7a976d11bd
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the string is a valid identifier
func IsIdentifier(s string) bool {
<<<<<<< HEAD
	return reValidIdentifier.MatchString(s)
=======
	return reValidName.MatchString(s)
>>>>>>> a486469478ac5f8553b60a97d8eb2a7a976d11bd
}
