package event

import (
	"context"
	"sync"

	// Package imports

	// Namespace imports
	"github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Receiver struct {
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return receiver as a string object
func (s *Receiver) String() string {
	str := "<event.receiver"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Receive events from one or more sources and process them using the callback. If
// a callback returns an error or the context is cancelled, then the function ends.
func (r *Receiver) Rcv(ctx context.Context, fn func(Event) error, src ...EventSource) error {
	var result error
	var wg sync.WaitGroup

	// Check parameters
	if len(src) == 0 || fn == nil {
		return nil
	}

	// Listen to all event sources
	for _, s := range src {
		wg.Add(1)
		go func(s EventSource) {
			defer wg.Done()
			if err := rcv(ctx, fn, s); err != nil {
				result = multierror.Append(result, err)
			}
		}(s)
	}

	// Wait for all goroutines to end
	wg.Wait()

	// Return any errors
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func rcv(ctx context.Context, fn func(Event) error, s EventSource) error {
	ch := s.Sub()
	defer s.Unsub(ch)

	// Loop until context cancelled, deadline exceeded or error returned
	// from callback function
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-ch:
			if err := fn(e); err != nil {
				return err
			}
		}
	}
}
