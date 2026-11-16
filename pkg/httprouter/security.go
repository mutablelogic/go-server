package httprouter

import (
	"bytes"
	"context"
	"net/http"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
)

///////////////////////////////////////////////////////////////////////////////
// INTERFACES

// SecurityScheme authenticates a request, such as by a bearer token or an
// API key, with the scopes an operation needs.
//
// An operation with more than one scheme, as alternatives, is run by a
// parent, which runs each scheme's Wrap at once, with a response of its own,
// and a handler which records the request it's called with. The first
// alternative, in the operation's order, which calls it, is the one whose
// request, with its context, the operation's handler has, and the others'
// contexts are cancelled. When none does, the response of the first which
// is forbidden, a 403, or else of the first, is the response. So a scheme:
//
//   - doesn't read the request's body, which the handler reads;
//   - doesn't write a response after it calls the handler;
//   - can run, and refuse, when another grants access, so its side effects,
//     such as counting failures, happen for each alternative;
//   - is safe to run concurrently, and returns when its context is
//     cancelled.
type SecurityScheme interface {
	// Wrap returns a HTTP handler function that wraps the security scheme around
	// the provided handler, enforcing the required scopes. It calls the
	// handler when the request has them, or else writes the response, such
	// as 401 when the request has no credentials, or 403 when they don't
	// have the scopes.
	Wrap(handler http.HandlerFunc, scopes []string) http.HandlerFunc

	// Spec returns the openapi.PathItem for a path with optional path parameters
	Spec() openapi.SecurityScheme
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterSecurityScheme registers a security scheme with the router. The scheme
// can then be referenced in the OpenAPI spec by name.
func (r *Router) RegisterSecurityScheme(name string, scheme SecurityScheme) error {
	if _, exists := r.security[name]; exists {
		return httpresponse.ErrConflict.Withf("security scheme %q already registered", name)
	}
	r.spec.AddSecurityScheme(name, scheme.Spec())
	r.security[name] = scheme
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE TYPES

// wrapFunc wraps a handler, such as with a security scheme
type wrapFunc func(http.HandlerFunc) http.HandlerFunc

// attempt is the outcome of an alternative: the request its schemes called
// the handler with, or nil when they didn't, and the response they wrote
type attempt struct {
	request  *http.Request
	response *recorder
}

// recorder is a response which a scheme writes, which is kept, rather than
// sent, until it's chosen
type recorder struct {
	header http.Header
	status int
	body   bytes.Buffer
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// allOf wraps a handler with each of the wraps, which are all required, the
// first outermost
func allOf(wraps []wrapFunc) wrapFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(wraps) - 1; i >= 0; i-- {
			next = wraps[i](next)
		}
		return next
	}
}

// anyOf wraps a handler with alternatives, any of which grants access. It
// runs them at once, each with a response of its own, and calls the handler
// with the request of the first, in order, which calls it, with the headers
// it set, and cancels the others. When none does, it writes the response of
// the first which is forbidden, or else of the first.
func anyOf(alternatives []wrapFunc) wrapFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Run the alternatives at once, each with a context of its own
			results := make([]chan attempt, len(alternatives))
			cancels := make([]context.CancelFunc, len(alternatives))
			for i, alternative := range alternatives {
				results[i] = make(chan attempt, 1)
				ctx, cancel := context.WithCancel(r.Context())
				cancels[i] = cancel
				go func(result chan<- attempt) {
					var granted *http.Request
					response := newRecorder()
					alternative(func(_ http.ResponseWriter, r *http.Request) {
						granted = r
					})(response, r.WithContext(ctx))
					result <- attempt{request: granted, response: response}
				}(results[i])
			}

			// The first, in order, which grants access, whose context lasts
			// until the handler returns, and the others are cancelled, at
			// once, as they may not return until they are
			var refused []attempt
			for i, result := range results {
				attempt := <-result
				if attempt.request == nil {
					cancels[i]()
					refused = append(refused, attempt)
					continue
				}
				defer cancels[i]()
				for _, cancel := range cancels[i+1:] {
					cancel()
				}
				copyHeader(w.Header(), attempt.response.header)
				next(w, attempt.request)
				return
			}

			// None did
			refusal(refused).writeTo(w)
		}
	}
}

// refusal returns the response of the first attempt which is forbidden, or
// else of the first
func refusal(attempts []attempt) *recorder {
	for _, attempt := range attempts {
		if attempt.response.status == http.StatusForbidden {
			return attempt.response
		}
	}
	return attempts[0].response
}

func newRecorder() *recorder {
	return &recorder{header: make(http.Header)}
}

func (r *recorder) Header() http.Header {
	return r.header
}

func (r *recorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

func (r *recorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(data)
}

// writeTo writes the response, which is unauthorized when a scheme wrote
// none, though it didn't call the handler
func (r *recorder) writeTo(w http.ResponseWriter) {
	if r.status == 0 {
		httpresponse.Error(w, httpresponse.ErrNotAuthorized)
		return
	}
	copyHeader(w.Header(), r.header)
	w.WriteHeader(r.status)
	_, _ = w.Write(r.body.Bytes())
}

// copyHeader copies the header's values to another
func copyHeader(to, from http.Header) {
	for key, values := range from {
		to[key] = append(to[key][:0:0], values...)
	}
}
