package httpauth

import (
	"context"
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AuthZ struct {
	// Set of policies, keyed by HTTP method, in uppercase
	Policies map[string][]*Policy
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	methodAny = "*"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAuthz() *AuthZ {
	return new(AuthZ)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE INTERFACE

func (a *AuthZ) WrapFunc(child http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Short circuit if no policies are defined
		if len(a.Policies) == 0 || (len(a.Policies[r.Method]) == 0 && len(a.Policies[methodAny]) == 0) {
			child(w, r)
			return
		}

		// Find first matching policy by method, then by wildcard
		ctx := r.Context()
		policy, matchedCtx := a.Match(a.Policies[r.Method], ctx, r)
		if policy == nil {
			policy, matchedCtx = a.Match(a.Policies[methodAny], ctx, r)
		}

		// No policy matched, deny by default
		if policy == nil {
			httpresponse.Error(w, httpresponse.ErrForbidden, r.URL.Path)
			return
		}

		// Apply the matched policy - allow or deny. It is the responsibility of the policy to write any error response when denying.
		if policy.Apply(w, r) {
			child(w, r.WithContext(matchedCtx))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (a *AuthZ) Match(policies []*Policy, ctx context.Context, r *http.Request) (*Policy, context.Context) {
	for _, policy := range policies {
		if matched := policy.Match(ctx, r); matched != nil {
			return policy, matched
		}
	}
	return nil, nil
}
