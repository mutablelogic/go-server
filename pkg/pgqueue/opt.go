package pgqueue

import (
	"strings"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	worker string
}

// Opt represents a function that modifies the options
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	var o opt

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

func OptWorker(worker string) Opt {
	return func(o *opt) error {
		if worker = strings.TrimSpace(worker); worker == "" {
			return httpresponse.ErrBadRequest.With("empty worker name")
		}
		o.worker = worker
		return nil
	}
}
