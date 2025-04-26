package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"

	// Packages
	"github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(ctx context.Context, jwt *Parser) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	googleKeysURL  = "https://www.googleapis.com/oauth2/v1/certs"
	cognitoKeysURL = "https://cognito-idp.${region}.amazonaws.com/${userPoolId}/.well-known/jwks.json"
	keyTypeRSA     = "RSA"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOptions(ctx context.Context, jwt *Parser, opts ...Opt) error {
	for _, opt := range opts {
		if err := opt(ctx, jwt); err != nil {
			return err
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithGoogle() Opt {
	return func(ctx context.Context, jwt *Parser) error {
		var response map[string]string
		if err := jwt.client.DoWithContext(ctx, client.MethodGet, &response, client.OptReqEndpoint(googleKeysURL)); err != nil {
			return err
		}

		// Decode the PEM blocks
		for kid, v := range response {
			block, rest := pem.Decode([]byte(v))
			if block == nil {
				return fmt.Errorf("failed to decode PEM block: %s", v)
			}
			if len(rest) > 0 {
				return fmt.Errorf("extra data found after PEM block: %s", rest)
			}
			if block.Type != "CERTIFICATE" {
				return fmt.Errorf("unexpected PEM block type: %s", block.Type)
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate: %s", err)
			}

			jwt.keys[kid] = key{
				Kid: kid,
				Key: cert.PublicKey,
			}
		}

		// Return success
		return nil
	}
}

func WithCognito(region, userPoolId string) Opt {
	return func(ctx context.Context, jwt *Parser) error {
		var response struct {
			Keys []key `json:"keys"`
		}
		if err := jwt.client.DoWithContext(ctx, client.MethodGet, &response, client.OptReqEndpoint(getCognitoKeysURL(region, userPoolId))); err != nil {
			return err
		}

		for _, k := range response.Keys {
			// We only support RSA keys for now
			if k.Type != keyTypeRSA {
				continue
			}

			// Decode exponent
			exponentBytes, err := base64.RawURLEncoding.DecodeString(k.E)
			if err != nil {
				return fmt.Errorf("failed to decode exponent: %w", err)
			}
			var exponent int
			if len(exponentBytes) < 8 { // Support up to 64 bits
				exponentBytes = append(make([]byte, 8-len(exponentBytes)), exponentBytes...) // Pad with leading zeros
				exponent = int(new(big.Int).SetBytes(exponentBytes).Uint64())
			} else {
				return fmt.Errorf("exponent too large: %s", k.E)
			}

			// Decode modulus
			modulusBytes, err := base64.RawURLEncoding.DecodeString(k.N)
			if err != nil {
				return fmt.Errorf("failed to decode modulus: %w", err)
			}
			modulus := new(big.Int).SetBytes(modulusBytes)

			// Create RSA public key
			k.Key = &rsa.PublicKey{
				N: modulus,
				E: exponent,
			}

			// Add key to the map
			jwt.keys[k.Kid] = k
		}

		// Return success
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func getCognitoKeysURL(region, userPoolId string) string {
	return os.Expand(cognitoKeysURL, func(key string) string {
		switch key {
		case "region":
			return region
		case "userPoolId":
			return userPoolId
		default:
			return ""
		}
	})
}
