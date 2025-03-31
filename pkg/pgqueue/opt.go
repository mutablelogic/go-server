package pgqueue

import (
	"fmt"
	"os"
	"strings"

	// Packages
	"github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	pg.OffsetLimit
	worker    string
	namespace string
}

// Opt represents a function that modifies the options
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	var o opt

	// Set the defaults
	o.namespace = schema.DefaultNamespace
	if hostname, err := os.Hostname(); err != nil {
		return nil, httpresponse.ErrInternalError.With(err)
	} else {
		o.worker = fmt.Sprint(hostname, ".", os.Getpid())
	}

	// Apply the options
	for _, fn := range opts {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}

	// Return success
	return &o, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set offset for the queue list
func OptOffset(offset uint64) Opt {
	return func(o *opt) error {
		o.Offset = offset
		return nil
	}
}

// Set limit for the queue list
func OptLimit(limit uint64) Opt {
	return func(o *opt) error {
		o.Limit = types.Uint64Ptr(limit)
		return nil
	}
}

// Set the worker name when a task is locked for work
func OptWorker(v string) Opt {
	return func(o *opt) error {
		if v = strings.TrimSpace(v); v == "" {
			return httpresponse.ErrBadRequest.With("empty worker name")
		} else {
			o.worker = v
		}
		return nil
	}
}

// Set the namespace for the tickers and queues
func OptNamespace(v string) Opt {
	return func(o *opt) error {
		if v = strings.TrimSpace(v); !types.IsIdentifier(v) {
			return httpresponse.ErrBadRequest.With("invalid namespacename ")
		} else {
			o.namespace = strings.ToLower(v)
		}
		return nil
	}
}
