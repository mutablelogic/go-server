package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type Token struct {
	Name   string    `json:"name,omitempty"`                            // Name of the token
	Value  string    `json:"token,omitempty"`                           // Token value
	Expire time.Time `json:"expire_time,omitempty"  writer:",width:29"` // Time of expiration for the token
	Time   time.Time `json:"access_time,omitempty" writer:",width:29"`  // Time of last access
	Scope  []string  `json:"scopes,omitempty" writer:",wrap"`           // Authentication scopes
}

type TokenCreate struct {
	Name     string   `json:"name,omitempty"`     // Name of the token
	Duration duration `json:"duration,omitempty"` // Duration of the token, or zero for no expiration
	Scope    []string `json:"scopes,omitempty"`   // Authentication scopes
}

type duration struct {
	time.Duration
}

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a token of the specified number of bytes, with the specified duration and scope.
// If the duration is zero, the token will not expire.
func NewToken(name string, length int, duration time.Duration, scope ...string) Token {
	var expire time.Time
	if duration != 0 {
		expire = time.Now().Add(duration)
	}
	return Token{
		Name:   name,
		Value:  generateToken(length),
		Time:   time.Now(),
		Scope:  scope,
		Expire: expire,
	}
}

// Create a new create token request
func NewCreateToken(name string, expires_in time.Duration, scope ...string) TokenCreate {
	// Truncase the duration to the nearest minute
	if expires_in > 0 {
		expires_in = expires_in.Truncate(time.Minute)
		if expires_in < time.Minute {
			expires_in = time.Minute
		}
	} else {
		expires_in = 0
	}

	// Return the token
	return TokenCreate{
		Name:     name,
		Duration: duration{expires_in},
		Scope:    scope,
	}
}

/////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Token) String() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Compares token name, value, expiry and scopes
func (t Token) Equals(other Token) bool {
	if t.Name != other.Name || t.Value != other.Value || t.Expire != other.Expire {
		return false
	}
	for _, scope := range other.Scope {
		if !slices.Contains(t.Scope, scope) {
			return false
		}
	}
	for _, scope := range t.Scope {
		if !slices.Contains(other.Scope, scope) {
			return false
		}
	}
	return true
}

// Return true if the token is valid (not expired)
func (t Token) IsValid() bool {
	if t.Expire.IsZero() || t.Expire.After(time.Now()) {
		return true
	}
	return false
}

// Return true if the token is a zero token
func (t Token) IsZero() bool {
	if t.Name == "" && t.Value == "" && t.Expire.IsZero() && t.Time.IsZero() && len(t.Scope) == 0 {
		return true
	}
	return false
}

// Return true if the token has the specified scope, and is valid
func (t Token) IsScope(scopes ...string) bool {
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

func (t Token) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

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

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value) * time.Second
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid duration of type %T", v)
	}
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
