package tokenauth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	// Package imports

	slices "golang.org/x/exp/slices"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	Name   string    `json:"name,omitempty"`        // Name of the token
	Value  string    `json:"token,omitempty"`       // Token value
	Expire time.Time `json:"expire_time,omitempty"` // Time of expiration for the token
	Time   time.Time `json:"access_time"`           // Time of last access
	Scope  []string  `json:"scopes,omitempty"`      // Authentication scopes
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
func (t *Token) IsScope(scopes ...string) bool {
	if !t.IsValid() {
		return false
	}
	for _, scope := range scopes {
		if slices.Contains(t.Scope, scope) {
			return true
		}
	}
	return false
}

/////////////////////////////////////////////////////////////////////
// JSON MARSHAL

func (t *Token) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if t == nil {
		return []byte("null"), nil
	}
	buf.WriteRune('{')

	// Write the fields
	if t.Name != "" {
		buf.WriteString(`"name":`)
		buf.WriteString(strconv.Quote(t.Name))
		buf.WriteRune(',')
	}
	if t.Value != "" {
		buf.WriteString(`"token":`)
		buf.WriteString(strconv.Quote(t.Value))
		buf.WriteRune(',')
	}
	if !t.Expire.IsZero() {
		buf.WriteString(`"expire_time":`)
		buf.WriteString(strconv.Quote(t.Expire.Format(time.RFC3339)))
		buf.WriteRune(',')
	}
	if !t.Time.IsZero() {
		buf.WriteString(`"access_time":`)
		buf.WriteString(strconv.Quote(t.Time.Format(time.RFC3339)))
		buf.WriteRune(',')
	}
	if len(t.Scope) > 0 {
		buf.WriteString(`"scopes":`)
		if data, err := json.Marshal(t.Scope); err != nil {
			return nil, err
		} else {
			buf.Write(data)
		}
		buf.WriteRune(',')
	}

	// Include the valid flag
	buf.WriteString(`"valid":`)
	buf.WriteString(strconv.FormatBool(t.IsValid()))

	// Return success
	buf.WriteRune('}')
	return buf.Bytes(), nil
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
