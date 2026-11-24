package otel

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
// names, which a handler marks, with RedactPathValue, and the middleware
// leaves out of its log, and its span
type redactions struct {
	mu     sync.Mutex
	values map[string]string // The values, by their names
}

// redactKey is the key of a request's redactions, in its context
type redactKey struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RedactPathValue marks the request's path values, by their names, such as
// a token in /hook/{token}, as secret, so HTTPHandler's log, and span, have
// the path with {token} in the value's place, including when the handler
// panics. A handler calls it, of a request which HTTPHandler is serving,
// before it does anything else. A name without a value is ignored, as is a
// request which HTTPHandler isn't serving. Only the path is redacted, rather
// than the query.
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

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// withRedactions returns the context, with redactions, which handlers mark
func withRedactions(ctx context.Context) context.Context {
	return context.WithValue(ctx, redactKey{}, &redactions{values: make(map[string]string)})
}

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
