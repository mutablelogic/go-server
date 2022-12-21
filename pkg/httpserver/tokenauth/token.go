package tokenauth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	// Package imports
	"golang.org/x/exp/slices"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	Value  string    `json:"token,omitempty"`       // Token value
	Expire time.Time `json:"expire_time,omitempty"` // Time of expiration for the token
	Time   time.Time `json:"access_time"`           // Time of last access
	Scope  []string  `json:"scopes"`                // Authentication scopes
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewToken(length int, duration time.Duration, scope ...string) *Token {
	var expire time.Time
	if duration != 0 {
		expire = time.Now().Add(duration)
	}
	return &Token{
		Value:  generateToken(length),
		Time:   time.Now(),
		Scope:  scope,
		Expire: expire,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t *Token) String() string {
	str := "<httpserver-token"
	str += fmt.Sprintf(" token=%q", t.Value)
	if !t.Time.IsZero() {
		str += fmt.Sprintf(" access_time=%q", t.Time.Format(time.RFC3339))
	}
	if !t.Expire.IsZero() {
		str += fmt.Sprintf(" expire_time=%q", t.Expire.Format(time.RFC3339))
	}
	if len(t.Scope) > 0 {
		str += fmt.Sprintf(" scopes=%q", t.Scope)
	}
	if t.IsValid() {
		str += " valid"
	}
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return true if the token is valid (not expired)
func (t *Token) IsValid() bool {
	if t.Expire.IsZero() || t.Expire.After(time.Now()) {
		return true
	}
	return false
}

// Return true if the token has the specified scope, and is valid
func (t *Token) IsScope(scope string) bool {
	if !t.IsValid() {
		return false
	}
	if slices.Contains(t.Scope, scope) {
		return true
	}
	return false
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
