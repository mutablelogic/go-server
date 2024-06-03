package tokenauth

type TokenJar interface {
	// Put a token into the jar. Update if the token exists, or create
	// a new token if it does not.
	Put(token *Token) *Token

	// Return a token from the jar, or nil if the token is not found.
	// The method should update the access time of the token.
	Get(token string) *Token

	// Remove a token from the jar. Return the token that was removed,
	// or nil if the token was not found.
	Remove(token string) *Token
}
