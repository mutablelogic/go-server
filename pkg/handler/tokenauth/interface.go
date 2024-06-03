package tokenauth

type TokenJar interface {
	// Return all tokens
	Tokens() []*Token

	// Return a token from the jar, or nil if the token is not found.
	// The method should update the access time of the token.
	Get(string) *Token

	// Put a token into the jar, assuming it does not yet exist.
	Create(*Token) error

	// Update an existing token in the jar, assuming it already exists.
	Update(*Token) error

	// Remove a token from the jar. Return the token that was removed,
	// or nil if the token was not found.
	Remove(string) error
}
