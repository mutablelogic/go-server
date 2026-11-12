// Package broadcaster publishes changes, such as to the rows of a server's
// tables, to the channels which watch it, so a server's clients see the
// changes as they happen
package broadcaster

import (
	"context"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Broadcaster sends the changes which are published to each channel which is
// watching it, whose filter they match. It's watched until it stops running.
type Broadcaster struct {
	mu       sync.Mutex
	stopped  bool
	channels map[chan<- schema.Change]*watch
}

// watch is a channel's watch: the changes it's sent, and when it ends
type watch struct {
	filter schema.ChangeFilter
	stop   chan struct{}
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// errStopped ends a watch, as the broadcaster has stopped running
var errStopped = httpresponse.ErrServiceUnavailable.With("broadcaster stopped")

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns a broadcaster, which can be watched
func New() *Broadcaster {
	return &Broadcaster{channels: make(map[chan<- schema.Change]*watch)}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run runs the broadcaster until the context is cancelled, when every watch
// ends, with ErrServiceUnavailable, as there will be no more events, such as
// when the server is stopping, and it isn't watched again until it runs
// again. It returns nil.
func (b *Broadcaster) Run(ctx context.Context) error {
	b.Start()
	<-ctx.Done()
	b.stop()
	return nil
}

// Start makes a broadcaster which has stopped running watched again, before
// it runs, so a watch made once Start returns isn't refused, as it would be
// while Run is still starting, such as in a goroutine. Run starts it too, so
// it's only needed by a caller which watches it as soon as it has started
// it, such as with:
//
//	b.Start()
//	go b.Run(ctx)
func (b *Broadcaster) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = false
}

// Watch sends the changes which are published, and match the filter, to the
// channel, which are all of them when the filter is empty, from when it
// returns, as it's watching before it does, until the context is cancelled,
// or the broadcaster stops running, as there will be no more changes.
//
// It returns a channel which is sent nil when the context is cancelled, or
// ErrServiceUnavailable when the broadcaster stops running, once the watch
// has ended, after which no more changes are sent to the channel, so it can
// be closed. A broadcaster which has stopped running isn't watched, and
// returns ErrServiceUnavailable, until it runs again.
//
// A change isn't waited for the channel to receive, so a channel which is
// full misses it, and one which needs every change should be buffered.
// Returns ErrBadRequest when there's no channel, or the filter's action
// isn't one of the actions, and ErrConflict when the channel is watching
// already.
func (b *Broadcaster) Watch(ctx context.Context, filter schema.ChangeFilter, ch chan<- schema.Change) (<-chan error, error) {
	if ch == nil {
		return nil, httpresponse.ErrBadRequest.With("missing channel")
	} else if err := filter.Validate(); err != nil {
		return nil, httpresponse.ErrBadRequest.With(err.Error())
	}

	// Add the channel, which is watching once at a time
	w := &watch{filter: filter, stop: make(chan struct{})}
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return nil, errStopped
	} else if _, exists := b.channels[ch]; exists {
		b.mu.Unlock()
		return nil, httpresponse.ErrConflict.With("channel is watching already")
	} else if b.channels == nil {
		b.channels = make(map[chan<- schema.Change]*watch)
	}
	b.channels[ch] = w
	b.mu.Unlock()

	// The watch ends when the context is cancelled, or the broadcaster
	// stops. It's removed before that's sent, under the lock events are sent
	// under, so once it's received, no more events are sent to the channel.
	done := make(chan error, 1)
	go func() {
		var err error
		select {
		case <-ctx.Done():
		case <-w.stop:
			err = errStopped
		}
		b.mu.Lock()
		if b.channels[ch] == w {
			delete(b.channels, ch)
		}
		b.mu.Unlock()
		done <- err
	}()
	return done, nil
}

// Publish sends the change to each channel which is watching, whose filter
// it matches, which can receive it now, and returns how many couldn't, as
// their channels were full, so they missed it. It never waits, so it can be
// called under a lock, which keeps the changes in the order they happened.
func (b *Broadcaster) Publish(change schema.Change) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	var missed int
	for ch, w := range b.channels {
		if !w.filter.Matches(change) {
			continue
		}
		select {
		case ch <- change:
		default:
			missed++
		}
	}
	return missed
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// stop ends every watch, as the broadcaster has stopped running, and those
// made later, until it runs again
func (b *Broadcaster) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = true
	for ch, w := range b.channels {
		close(w.stop)
		delete(b.channels, ch)
	}
}
