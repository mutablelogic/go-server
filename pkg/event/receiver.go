package event

import (
	"context"
	"sync"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Receiver struct {
}

// Compile time check
var _ EventReceiver = (*Receiver)(nil)

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
			ch := s.Sub()
			defer s.Unsub(ch)
		FOR_LOOP:
			for {
				select {
				case <-ctx.Done():
					if err := ctx.Err(); err != nil {
						result = multierror.Append(err)
					}
					break FOR_LOOP
				case e := <-ch:
					if err := fn(e); err != nil {
						result = multierror.Append(err)
						break FOR_LOOP
					}
				}
			}
		}(s)
	}

	// Wait for all goroutines to end
	wg.Wait()

	// Return any errors
	return result
}
