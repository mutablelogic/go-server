/*
implements a token jar that stores tokens into memory, and potentially a file
on the file system
*/
package tokenjar

////////////////////////////////////////////////////////////////////////////////
// TYPES

type tokenjar struct {
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new tokenjar, with the specified path. If the path is empty,
// the tokenjar will be in-memory only.
func New(path string) (*tokenjar, error) {
	j := new(tokenjar)

	// Return success
	return j, nil
}
