package auth

import (
	"context"
)

type TokenJar interface {
	server.Task

	// Return all tokens
	Tokens() []Token

	// Return a token from the jar by value, or an invalid token
	// if the token is not found. The method should update the access
	// time of the token.
	GetWithValue(string) Token

	// Return a token from the jar by name, or nil if the token
	// is not found. The method should not update the access time
	// of the token.
	GetWithName(string) Token

	// Put a token into the jar, assuming it does not yet exist.
	Create(Token) error

	// Update an existing token in the jar, assuming it already exists.
	Update(Token) error

	// Remove a token from the jar, based on key.
	Delete(string) error
}
