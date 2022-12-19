package tokenauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Module imports
	event "github.com/mutablelogic/go-server/pkg/event"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type tokenauth struct {
	sync.RWMutex
	task.Task

	delta    time.Duration
	path     string
	tokens   map[string]*Token
	modified bool
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new logger task with provider of other tasks
func NewWithPlugin(p Plugin) (*tokenauth, error) {
	this := new(tokenauth)

	// Construct the path to the file
	if path, err := p.Path(); err != nil {
		return nil, err
	} else if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, ErrBadParameter.Withf("not a directory: %q", path)
	} else if path, err := filepath.Abs(filepath.Join(path, p.File())); err != nil {
		return nil, err
	} else {
		this.path = path
		this.delta = p.Delta()
	}

	// Read the file if it exists
	if tokens, err := fileRead(this.path); err != nil {
		return nil, err
	} else {
		this.tokens = tokens
	}

	// If the admin token does not exist, then create it
	if _, ok := this.tokens[AdminToken]; !ok {
		// Create a new token
		this.tokens[AdminToken] = newToken(defaultLength)
	}

	// Write tokens to disk
	if err := fileWrite(this.path, this.tokens); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

// Run will write the authorization tokens back to disk if they have been modified
func (tokenauth *tokenauth) Run(ctx context.Context) error {
	ticker := time.NewTicker(tokenauth.delta)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_, err := tokenauth.writeIfModified()
			return err
		case <-ticker.C:
			tokenauth.Lock()
			if written, err := tokenauth.writeIfModified(); err != nil {
				tokenauth.Emit(event.Error(ctx, err))
			} else if written {
				tokenauth.Emit(event.New(ctx, EventTypeWrite, "Written tokens to disk"))
			}
			tokenauth.Unlock()
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (tokenauth *tokenauth) String() string {
	str := "<httpserver-tokenauth"
	str += fmt.Sprint(" delta=", tokenauth.delta)
	str += fmt.Sprintf(" path=%q", tokenauth.path)
	str += fmt.Sprintf(" tokens=%d", len(tokenauth.tokens))
	if tokenauth.modified {
		str += " modified"
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if a token associated with the name already exists
func (tokenauth *tokenauth) Exists(name string) bool {
	tokenauth.RLock()
	defer tokenauth.RUnlock()

	_, ok := tokenauth.tokens[name]
	return ok
}

// Create a new token associated with a name and return it.
func (tokenauth *tokenauth) Create(name string) (string, error) {
	tokenauth.Lock()
	defer tokenauth.Unlock()

	// If the name is invalid, then return an error
	if !types.IsIdentifier(name) {
		return "", ErrBadParameter.Withf("%q", name)
	}
	// If the name exists already, then return an error
	if _, ok := tokenauth.tokens[name]; ok {
		return "", ErrDuplicateEntry.Withf("%q", name)
	}
	// If the name is the admin token, then return an error
	if name == AdminToken {
		return "", ErrBadParameter.Withf("%q", name)
	}

	// Create a new token
	tokenauth.tokens[name] = newToken(defaultLength)

	// Set modified flag
	tokenauth.setModified(true)

	// Success: return the token value
	return tokenauth.tokens[name].Value, nil
}

// Revoke a token associated with a name. For the admin token, it is
// rotated rather than revoked.
func (tokenauth *tokenauth) Revoke(name string) error {
	tokenauth.Lock()
	defer tokenauth.Unlock()

	// If the name does not exist, then return an error
	if _, ok := tokenauth.tokens[name]; !ok {
		return ErrNotFound.Withf("%q", name)
	}

	// Either delete or rotate the token
	var immediately bool
	if name == AdminToken {
		// Rotate the token
		tokenauth.tokens[name] = newToken(defaultLength)
		// Write immediately
		immediately = true
	} else {
		// Delete the token
		delete(tokenauth.tokens, name)
	}

	// Set modified flag
	tokenauth.setModified(true)

	// Write to disk immediately when admin token is rotated
	if immediately {
		if written, err := tokenauth.writeIfModified(); err != nil {
			return err
		} else if written {
			// TODO: Rotated admin token
			return nil
		}
	}

	// Return success
	return nil
}

// Return all token names and their last access times
func (tokenauth *tokenauth) Enumerate() map[string]time.Time {
	tokenauth.RLock()
	defer tokenauth.RUnlock()

	var result = make(map[string]time.Time)
	for k, v := range tokenauth.tokens {
		result[k] = v.Time
	}

	// Return the result
	return result
}

// Returns the name of the token if a value matches. Updates
// the access time for the token. If token with value not
// found, then return empty string
func (tokenauth *tokenauth) Matches(value string) string {
	tokenauth.Lock()
	defer tokenauth.Unlock()

	for k, v := range tokenauth.tokens {
		if v.Value == value {
			v.Time = time.Now()
			tokenauth.setModified(true)
			return k
		}
	}

	// Token not found
	return ""
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// setModified sets a new modified value, and returns true if changed
func (tokenauth *tokenauth) setModified(modified bool) bool {
	if modified != tokenauth.modified {
		tokenauth.modified = modified
		return true
	} else {
		return false
	}
}

// write the tokens to disk if modified
func (tokenauth *tokenauth) writeIfModified() (bool, error) {
	modified := tokenauth.setModified(false)
	if modified {
		if err := fileWrite(tokenauth.path, tokenauth.tokens); err != nil {
			return modified, err
		}
	}

	// Return success
	return modified, nil
}

func fileRead(filename string) (map[string]*Token, error) {
	var result = map[string]*Token{}

	// If the file doesn't exist, return empty result
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return result, nil
	} else if err != nil {
		return nil, err
	}

	// Open the file
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Decode the file
	if err := json.NewDecoder(fh).Decode(&result); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

func fileWrite(filename string, tokens map[string]*Token) error {
	if tokens == nil {
		return ErrBadParameter.Withf("tokens is nil")
	}

	// Create the file
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Write the tokens
	if err := json.NewEncoder(fh).Encode(tokens); err != nil {
		return err
	}

	// Return success
	return nil
}

func newToken(length int) *Token {
	return &Token{
		Value: generateToken(length),
		Time:  time.Now(),
	}
}

func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
