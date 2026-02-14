package httpauth

import (
	"context"
	"net/http"
	"strings"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// CONTEXT KEYS

type contextKey int

const (
	contextKeyToken contextKey = iota
	contextKeyUsername
	contextKeyPassword
	contextKeyApiKey
	contextKeyDigest
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Action represents the policy decision
type Action int
type Auth string

const (
	Allow Action = iota
	Deny
)

const (
	None   Auth = "none"
	Bearer Auth = "bearer"
	Basic  Auth = "basic"
	Digest Auth = "digest"
	ApiKey Auth = "apikey"
)

type Policy struct {
	Action  Action
	Methods []string
	Auth    Auth // bearer, basic, digest, apikey
	Expr    []string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPolicy() *Policy {
	return new(Policy)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Match returns a new context with auth values if the policy conditions match
// the request, or nil if the policy does not match.
func (p *Policy) Match(ctx context.Context, r *http.Request) context.Context {
	// Check Auth
	switch p.Auth {
	case None:
		// No authentication required, allow if other conditions match
	case Bearer:
		if ctx = matchAuthBearer(ctx, r); ctx == nil {
			return nil
		}
	case Basic:
		if ctx = matchAuthBasic(ctx, r); ctx == nil {
			return nil
		}
	case Digest:
		if ctx = matchAuthDigest(ctx, r); ctx == nil {
			return nil
		}
	case ApiKey:
		if ctx = matchAuthApiKey(ctx, r); ctx == nil {
			return nil
		}
	default:
		return nil
	}

	// TODO: Check expression conditions
	return ctx
}

// Apply enforces the policy decision. Returns true if the request is allowed
// to proceed, or false if it was denied (and an error response was written).
func (p *Policy) Apply(w http.ResponseWriter, r *http.Request) bool {
	if p.Action == Deny {
		httpresponse.Error(w, httpresponse.ErrForbidden, r.URL.Path)
		return false
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func matchAuthBearer(ctx context.Context, r *http.Request) context.Context {
	if scheme, value := authValues(r); scheme == Bearer && value != "" {
		return context.WithValue(ctx, contextKeyToken, value)
	}
	return nil
}

func matchAuthBasic(ctx context.Context, r *http.Request) context.Context {
	username, password, ok := r.BasicAuth()
	if !ok || username == "" || password == "" {
		return nil
	}
	ctx = context.WithValue(ctx, contextKeyUsername, username)
	ctx = context.WithValue(ctx, contextKeyPassword, password)
	return ctx
}

func matchAuthDigest(ctx context.Context, r *http.Request) context.Context {
	if scheme, value := authValues(r); scheme == Digest && value != "" {
		// Check for required digest fields
		for _, field := range []string{"username=", "realm=", "nonce=", "response="} {
			if !strings.Contains(value, field) {
				return nil
			}
		}
		return context.WithValue(ctx, contextKeyDigest, value)
	}
	return nil
}

func matchAuthApiKey(ctx context.Context, r *http.Request) context.Context {
	header := r.Header.Get(types.ApiKeyHeader)
	if header == "" {
		// Fall back to Authorization header with "apikey" scheme
		if scheme, value := authValues(r); scheme == ApiKey {
			header = value
		}
	}

	// X-API-Key can be in header or in Authorization header
	if header == "" {
		return nil
	}
	return context.WithValue(ctx, contextKeyApiKey, header)
}

func authValues(r *http.Request) (Auth, string) {
	value := r.Header.Get(types.AuthorizationHeader)
	if parts := strings.SplitN(value, " ", 2); len(parts) == 2 {
		return Auth(strings.ToLower(strings.TrimSpace(parts[0]))), strings.TrimSpace(parts[1])
	} else {
		return "", ""
	}
}
