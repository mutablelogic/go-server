package event

import (
	"context"

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

func (s *Receiver) String() string {
	str := "<event.receiver"
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *Receiver) Rcv(ctx context.Context, fn func(Event) error, src ...EventSource) error {
	var result error
	if len(src) == 0 {
		return nil
	}
	for _, s := range src {
		go r.rcv(ctx, fn, s)
	}
	return result
}
