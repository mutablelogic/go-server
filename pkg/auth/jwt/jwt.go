package jwt

import (
	"context"
	"os"

	// Packages
	jwt "github.com/golang-jwt/jwt/v5"
	client "github.com/mutablelogic/go-client"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Parser struct {
	client *client.Client
	keys   map[string]key
}

type key struct {
	Kid  string `json:"kid"` // Key ID
	Type string `json:"kty"` // Public Key Type
	Alg  string `json:"alg"` // Signing Algorithm
	E    string `json:"e"`   // Exponent
	N    string `json:"n"`   // Modulus
	Key  any    `json:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ctx context.Context, opt ...Opt) (*Parser, error) {
	parser := new(Parser)
	parser.keys = make(map[string]key, 10)

	// Create a new client
	if client, err := client.New(client.OptEndpoint("http://localhost/"), client.OptTrace(os.Stderr, true)); err != nil {
		return nil, err
	} else {
		parser.client = client
	}

	// Apply options
	if err := applyOptions(ctx, parser, opt...); err != nil {
		return nil, err
	}

	// Return success
	return parser, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (parser *Parser) Decode(value string) (*jwt.Token, error) {
	return jwt.Parse(value, func(token *jwt.Token) (any, error) {
		key, err := parser.keyForId(token.Header["kid"])
		if err != nil {
			return nil, err
		}
		return key.Key, nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (parser *Parser) keyForId(kid any) (*key, error) {
	if kid_str, ok := kid.(string); !ok {
		return nil, httpresponse.ErrNotAuthorized.Withf("invalid kid: %q", kid)
	} else if key, ok := parser.keys[kid_str]; ok {
		return &key, nil
	}
	return nil, httpresponse.ErrNotAuthorized.Withf("unknown kid: %q", kid)
}
