package server

import "context"

/////////////////////////////////////////////////////////////////////
// TEMPLATE & INDEXER INTERFACES

// Env interface returns an environment variable
type Env interface {
	// GetString returns a string value for key, or ErrNotFound
	GetString(context.Context, string) (string, error)
}
