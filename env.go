package server

/////////////////////////////////////////////////////////////////////
// TEMPLATE & INDEXER INTERFACES

// Env interface returns an environment variable
type Env interface {
	// GetString returns a string value for key, or ErrNotFound
	GetString(string) (string, error)
}
