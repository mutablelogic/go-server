/*
implements a token jar that stores tokens into memory, and potentially a file
on the file system
*/
package tokenjar

import (
	"sync"
	"time"

	// Package imports
	"github.com/mutablelogic/go-server/pkg/handler/tokenauth"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type tokenjar struct {
	sync.RWMutex
	jar map[string]*tokenauth.Token
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCap = 20
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new tokenjar, with the specified path. If the path is empty,
// the tokenjar will be in-memory only.
func New(path string) (*tokenjar, error) {
	j := new(tokenjar)

	// Create the token jar
	j.jar = make(map[string]*tokenauth.Token, defaultCap)

	// Return success
	return j, nil
}

// Return all tokens
func (jar *tokenjar) Tokens() []*tokenauth.Token {
	var result []*tokenauth.Token

	// Lock the jar for read
	jar.RLock()
	defer jar.RUnlock()

	// Copy the tokens
	for _, token := range jar.jar {
		result = append(result, token)
	}

	// Return the result
	return result
}

// Return a token from the jar, or nil if the token is not found.
// The method should update the access time of the token.
func (jar *tokenjar) Get(key string) *tokenauth.Token {
	jar.Lock()
	defer jar.Unlock()

	if token, ok := jar.jar[key]; ok {
		token.Time = time.Now()
		return token
	} else {
		return nil
	}
}

// Put a token into the jar, assuming it does not yet exist.
func (jar *tokenjar) Create(token *tokenauth.Token) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if token == nil || token.Value == "" {
		return ErrBadParameter
	}
	if _, ok := jar.jar[token.Value]; ok {
		return ErrDuplicateEntry
	}

	// Update the token
	token.Time = time.Now()
	jar.jar[token.Value] = token

	// Return success
	return nil
}

// Update an existing token in the jar, assuming it already exists.
func (jar *tokenjar) Update(token *tokenauth.Token) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if token == nil || token.Value == "" {
		return ErrBadParameter
	}
	if _, ok := jar.jar[token.Value]; !ok {
		return ErrNotFound
	}

	// Update the token
	token.Time = time.Now()
	jar.jar[token.Value] = token

	// Return success
	return nil
}

// Remove a token from the jar. Return the token that was removed,
// or nil if the token was not found.
func (jar *tokenjar) Remove(key string) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if _, ok := jar.jar[key]; !ok {
		return ErrNotFound
	} else {
		delete(jar.jar, key)
	}

	// Return success
	return nil
}
