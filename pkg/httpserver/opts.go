package httpserver

import (
	"net/http"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	r        time.Duration
	w        time.Duration
	i        time.Duration
	catchAll http.Handler
}

type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opts ...Opt) (*opt, error) {
	o := new(opt)
	o.r = defaultReadTimeout
	o.w = defaultWriteTimeout
	o.i = defaultIdleTimeout
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set the read timeout for the HTTP server
func WithReadTimeout(v time.Duration) Opt {
	return func(o *opt) error {
		if v > 0 {
			o.r = v
		}
		return nil
	}
}

// Set the write timeout for the HTTP server
func WithWriteTimeout(v time.Duration) Opt {
	return func(o *opt) error {
		if v > 0 {
			o.w = v
		}
		return nil
	}
}

// Set the idle timeout for keep-alive connections
func WithIdleTimeout(v time.Duration) Opt {
	return func(o *opt) error {
		if v > 0 {
			o.i = v
		}
		return nil
	}
}

// WithCatchAll wraps the server's router so that any request not matched by the
// router is handled by h instead of returning Go's default plain-text 404.
func WithCatchAll(h http.Handler) Opt {
	return func(o *opt) error {
		o.catchAll = h
		return nil
	}
}
