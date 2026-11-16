package httprouter

import (
	"bytes"
	"context"
	"net/http"
	"sync"

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
// parent, which runs each scheme's Wrap at once, with a clone of the
// request, a response of its own, and a handler which waits. The first
// alternative, in the operation's order, which calls it, is the one whose
// request, with its context, the operation's handler has, while it waits,
// so its deferred cleanup runs after the handler. The others' contexts are
// cancelled. When none does, the response of the first which is forbidden,
// a 403, or else of the first, is the response. A panic in a scheme which
// is waited for is the request's, as it is for one scheme, and in another
// it's dropped. So a scheme:
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
	response *recorder
	panic    any
}

// runner is an alternative which runs in a goroutine of its own: it's
// granted, with the request its schemes called the handler with, or it
// returns, with what it wrote, or its panic, and it's released, once it's
// granted, when its handler has run, or it's lost
type runner struct {
	granted  chan *http.Request
	done     chan attempt
	release  func()
	cancel   context.CancelFunc
	response func() http.Header
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
// runs them at once, each with a clone of the request, and a response of
// its own, and calls the handler with the request of the first, in order,
// which calls it, with the headers it set, while that alternative waits,
// so its deferred cleanup runs after the handler. The others are cancelled,
// and released. When none does, it writes the response of the first which
// is forbidden, or else of the first. A panic in an alternative which is
// waited for is the request's, as one in a scheme which wraps the handler
// is, and others are dropped.
func anyOf(alternatives []wrapFunc) wrapFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			runners := make([]*runner, len(alternatives))
			for i, alternative := range alternatives {
				runners[i] = run(r, alternative)
			}

			// stop cancels and releases the alternatives, other than one
			stop := func(except int) {
				for i, runner := range runners {
					if i != except {
						runner.cancel()
						runner.release()
					}
				}
			}

			// The first, in order, which grants access, or else what each
			// wrote, until one panics
			var refused []attempt
			for i, runner := range runners {
				select {
				case request := <-runner.granted:
					stop(i)
					defer runner.cancel()
					copyHeader(w.Header(), runner.response())
					func() {
						defer runner.release()
						next(w, request)
					}()

					// Its cleanup, after the handler, which can panic
					if attempt := <-runner.done; attempt.panic != nil {
						panic(attempt.panic)
					}
					return
				case attempt := <-runner.done:
					if attempt.panic != nil {
						stop(-1)
						panic(attempt.panic)
					}
					runner.cancel()
					refused = append(refused, attempt)
				}
			}

			// None did
			refusal(refused).writeTo(w)
		}
	}
}

// run runs the alternative, in a goroutine of its own, with a clone of the
// request, whose context is cancelled by the runner, and a response of its
// own. When its schemes call the handler, it's granted, and waits, until
// it's released. Its panic is recovered, and returned.
func run(r *http.Request, alternative wrapFunc) *runner {
	ctx, cancel := context.WithCancel(r.Context())
	released := make(chan struct{})
	runner := &runner{
		granted: make(chan *http.Request, 1),
		done:    make(chan attempt, 1),
		release: sync.OnceFunc(func() { close(released) }),
		cancel:  cancel,
	}
	response := newRecorder()
	runner.response = func() http.Header { return response.header }
	go func() {
		var attempt attempt
		defer func() {
			attempt.panic = recover()
			attempt.response = response
			runner.done <- attempt
		}()
		var once sync.Once
		alternative(func(_ http.ResponseWriter, r *http.Request) {
			// Granted, once, and wait, until the handler has run, or it's
			// lost
			once.Do(func() {
				runner.granted <- r
				<-released
			})
		})(response, r.Clone(ctx))
	}()
	return runner
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
