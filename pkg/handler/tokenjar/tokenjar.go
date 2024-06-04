/*
implements a token jar that stores tokens into memory, and potentially a file
on the file system
*/
package tokenjar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Package imports
	server "github.com/mutablelogic/go-server"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type tokenjar struct {
	sync.RWMutex

	// Set the write interval to persist the tokens to disk
	writeInterval time.Duration

	// The filename to persist the tokens to
	filename string

	// Tokens keyed by the token value
	jar map[string]*auth.Token

	// Modified flag when the persistenr storage is updated
	modified bool
}

var _ auth.TokenJar = (*tokenjar)(nil)
var _ server.Task = (*tokenjar)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCap           = 20
	defaultFilename      = "tokenauth.json"
	defaultWriteInterval = time.Minute
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new tokenjar, with the specified path. If the path is empty,
// the tokenjar will be in-memory only.
func New(c Config) (*tokenjar, error) {
	j := new(tokenjar)

	// Set filepath for persistent storage
	if c.DataPath != "" {
		if stat, err := os.Stat(c.DataPath); err != nil {
			return nil, err
		} else if !stat.IsDir() {
			return nil, ErrBadParameter.Withf("not a directory: %v", c.DataPath)
		} else {
			j.filename = filepath.Join(c.DataPath, defaultFilename)
		}
	}

	// Set write interval
	if c.WriteInterval == 0 {
		j.writeInterval = defaultWriteInterval
	} else {
		j.writeInterval = c.WriteInterval
	}

	// Read the tokens from persistent storage
	var tokens []*auth.Token
	if _, err := os.Stat(j.filename); os.IsNotExist(err) {
		// Do nothing
	} else if err != nil {
		return nil, err
	} else if tokens_, err := j.Read(); err != nil {
		return nil, err
	} else {
		tokens = tokens_
	}

	// Create the token jar
	j.jar = make(map[string]*auth.Token, len(tokens)+defaultCap)

	// Read persistent tokens, bail if there is an inconsistent file
	for _, token := range tokens {
		if _, exists := j.jar[token.Value]; exists {
			return nil, ErrDuplicateEntry.With(token.Value)
		}
		if token.IsValid() {
			j.jar[token.Value] = token
		} else {
			j.modified = true
		}
	}

	// Return success
	return j, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the jar has been modified
func (jar *tokenjar) Modified() bool {
	jar.RLock()
	defer jar.RUnlock()

	return jar.modified
}

// Return all tokens
func (jar *tokenjar) Tokens() []auth.Token {
	var result []auth.Token

	// Lock the jar for read
	jar.RLock()
	defer jar.RUnlock()

	// Copy the tokens
	for _, token := range jar.jar {
		result = append(result, *token)
	}

	// Return the result
	return result
}

// Return a token from the jar, or nil if the token is not found.
// The method should update the access time of the token.
func (jar *tokenjar) Get(key string) auth.Token {
	jar.Lock()
	defer jar.Unlock()

	if token, ok := jar.jar[key]; ok {
		token.Time = time.Now()
		jar.modified = true

		// Make a copy of the token before returning
		return *token
	} else {
		// Return an empty token - not found
		return auth.Token{}
	}
}

// Put a token into the jar, assuming it does not yet exist.
func (jar *tokenjar) Create(token auth.Token) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if token.Value == "" {
		return ErrBadParameter
	}
	if _, ok := jar.jar[token.Value]; ok {
		return ErrDuplicateEntry
	}

	// Update the token
	token.Time = time.Now()
	jar.jar[token.Value] = &token
	jar.modified = true

	// Return success
	return nil
}

// Update an existing token in the jar, assuming it already exists.
func (jar *tokenjar) Update(token auth.Token) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if token.Value == "" {
		return ErrBadParameter
	}
	dest, ok := jar.jar[token.Value]
	if !ok {
		return ErrNotFound
	}

	// Update the token
	dest.Name = token.Name
	dest.Time = time.Now()
	dest.Expire = token.Expire
	dest.Scope = append([]string{}, token.Scope...)
	jar.modified = true

	// Return success
	return nil
}

// Remove a token from the jar
func (jar *tokenjar) Delete(key string) error {
	jar.Lock()
	defer jar.Unlock()

	// Check if the token already exists
	if _, ok := jar.jar[key]; !ok {
		return ErrNotFound
	} else {
		delete(jar.jar, key)
		jar.modified = true
	}

	// Return success
	return nil
}

// Write the tokens to persistent storage
func (jar *tokenjar) Write() error {
	jar.Lock()
	defer jar.Unlock()

	// NOP if there is no filename
	if jar.filename == "" {
		return nil
	}

	// Open the file for writing
	w, err := os.Create(jar.filename)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write the tokens, unset modified flag
	jar.modified = false
	if err := json.NewEncoder(w).Encode(jar.jar); err != nil {
		return err
	}

	// Return success
	return nil
}

// Run the token jar
func (jar *tokenjar) Read() ([]*auth.Token, error) {
	// NOP if there is no filename
	if jar.filename == "" {
		return nil, nil
	}

	// Open the file for reading
	r, err := os.Open(jar.filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Read the tokens
	var tokens []*auth.Token
	if err := json.NewDecoder(r).Decode(&tokens); err != nil {
		return nil, err
	}

	// Return success
	return tokens, nil
}
