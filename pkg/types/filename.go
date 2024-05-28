package types

import "regexp"

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// An file is a string which starts with a letter or number and is followed by
	// letters, numbers, underscores or hyphens and punkts. If needs to be at least
	// 2 characters long. If cannot end with a punkt, underscore or hyphen.
	ReFilename = `[a-zA-Z0-9][a-zA-Z0-9_\-\.]*[a-zA-Z0-9]`
)

var (
	reValidFilename = regexp.MustCompile(`^` + ReFilename + `$`)
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the string is a valid filename
func IsFilename(s string) bool {
	return reValidFilename.MatchString(s)
}
