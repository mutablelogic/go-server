package pgqueue

import (
	"strings"

	// Packages

	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
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

// Set the worker name when a task is locked for work
func OptWorker(v string) Opt {
	return func(o *opt) error {
		if v = strings.TrimSpace(v); v != "" {
			o.worker = v
		}
		return nil
	}
}

// Set the namespace for the tickers and queues
func OptNamespace(v string) Opt {
	return func(o *opt) error {
		if v = strings.TrimSpace(v); !types.IsIdentifier(v) || v == schema.DefaultNamespace || v == schema.CleanupNamespace {
			return httpresponse.ErrBadRequest.With("invalid namespace")
		} else {
			o.namespace = strings.ToLower(v)
		}
		return nil
	}
}
