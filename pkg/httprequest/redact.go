package httprequest

import (
	"cmp"
	"context"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// redactions are the path values of a request which are secret, by their
// names, which are left out of what's logged, and traced, of it
type redactions struct {
	mu     sync.Mutex
	values map[string]string // The values, by their names
}

// redactKey is the key of a request's redactions, in its context
type redactKey struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Redact returns the handler, whose requests' path values of the names, such
// as a token in /hook/{token}, are secret, and whose requests can have more
// marked, with RedactPathValue. A router wraps the middleware of a route
// with it, whose path parameters are secret, so they're secret whichever
// handler, or middleware, answers.
func Redact(next http.HandlerFunc, names ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		served := r
		if redactionsFrom(r.Context()) == nil {
			served = r.WithContext(context.WithValue(r.Context(), redactKey{}, &redactions{values: make(map[string]string)}))

			// The server removes the files of a request's form, of the
			// request which it made
			defer func() {
				if served.MultipartForm != nil {
					r.MultipartForm = served.MultipartForm
				}
			}()
		}
		RedactPathValue(served, names...)
		next(w, served)
	}
}

// RedactPathValue marks the request's path values, by their names, as
// secret, so they're left out of what's logged, and traced, of it. A name
// without a value is ignored, as is a request which Redact isn't serving.
// Only the path is redacted, rather than the query.
func RedactPathValue(r *http.Request, names ...string) {
	red := redactionsFrom(r.Context())
	if red == nil {
		return
	}
	red.mu.Lock()
	defer red.mu.Unlock()
	for _, name := range names {
		if value := r.PathValue(name); value != "" {
			red.values[name] = value
		}
	}
}

// RedactedURL returns the path of the URL, and the URL, of the context's
// request, as they can be logged, with the name, in braces, in place of each
// path value which is secret, and whether there were any
func RedactedURL(ctx context.Context, u *url.URL) (path, full string, redacted bool) {
	return redactionsFrom(ctx).redact(u)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// redactionsFrom returns the redactions of the context, or nil
func redactionsFrom(ctx context.Context) *redactions {
	red, _ := ctx.Value(redactKey{}).(*redactions)
	return red
}

// redact returns the URL's path, and the URL, as they're logged, and whether
// they were redacted: each value which is marked is replaced by its name, in
// braces, where it's the whole of a segment of the path, or of more than
// one, as the mux matches it, once the segments are unescaped. The rest of
// the URL's path is as it was escaped.
func (red *redactions) redact(u *url.URL) (path, full string, redacted bool) {
	if red == nil {
		return u.Path, u.String(), false
	}
	red.mu.Lock()
	defer red.mu.Unlock()
	if len(red.values) == 0 {
		return u.Path, u.String(), false
	}

	// The longest first, so a value isn't taken to be one which starts it
	names := make([]string, 0, len(red.values))
	for name := range red.values {
		names = append(names, name)
	}
	slices.SortFunc(names, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(red.values[b]), len(red.values[a])), cmp.Compare(a, b))
	})

	// The segments, as they're escaped, and unescaped
	escaped := strings.Split(u.EscapedPath(), "/")
	decoded := make([]string, len(escaped))
	for i, segment := range escaped {
		if value, err := url.PathUnescape(segment); err == nil {
			decoded[i] = value
		} else {
			decoded[i] = segment
		}
	}

	// Replace the values, once each segment, from the first
	var pathOut, fullOut []string
	for i := 0; i < len(decoded); {
		name, n := red.match(names, decoded[i:])
		if n == 0 {
			pathOut, fullOut = append(pathOut, decoded[i]), append(fullOut, escaped[i])
			i++
			continue
		}
		pathOut, fullOut = append(pathOut, "{"+name+"}"), append(fullOut, "{"+name+"}")
		redacted = true
		i += n
	}

	// The URL, which is the path, and the query, of a server's request,
	// unless it's a proxy's, which has a host
	full = strings.Join(fullOut, "/")
	if u.Scheme != "" || u.Host != "" {
		full = (&url.URL{Scheme: u.Scheme, User: u.User, Host: u.Host}).String() + full
	}
	if u.RawQuery != "" || u.ForceQuery {
		full += "?" + u.RawQuery
	}
	return strings.Join(pathOut, "/"), full, redacted
}

// match returns the name of the value which the segments start with, and how
// many of them it is, or zero, when none is
func (red *redactions) match(names, segments []string) (string, int) {
	for _, name := range names {
		value := red.values[name]
		for n := 1; n <= len(segments); n++ {
			joined := strings.Join(segments[:n], "/")
			if joined == value {
				return name, n
			} else if len(joined) > len(value) {
				break
			}
		}
	}
	return "", 0
}
