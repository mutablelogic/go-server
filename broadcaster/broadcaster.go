// Package broadcaster publishes changes, such as to the rows of a server's
// tables, to the channels which watch it, so a server's clients see the
// changes as they happen
package broadcaster

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

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
	running  atomic.Bool
	channels map[chan<- schema.Change]*watch
}

// watch is a channel's watch: the changes it's sent, when it ends, and the
// tables it's owed a rescan of, as it missed a change of them
type watch struct {
	filter schema.ChangeFilter
	stop   chan struct{}
	owed   []table        // The tables it's owed a rescan of, in the order it missed them
	owes   map[table]bool // Whether it's owed a rescan of a table
}

// table is a table of a schema, which a rescan is of
type table struct {
	schema, name string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// errStopped ends a watch, as the broadcaster has stopped running
var errStopped = httpresponse.ErrServiceUnavailable.With("broadcaster stopped")

// RescanInterval is how often Run sends the rescans which watches are owed,
// when no changes are published to send them with
var RescanInterval = 100 * time.Millisecond

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
// again. While it runs, it sends the rescans which watches are owed, at the
// RescanInterval (see Flush). It returns nil.
func (b *Broadcaster) Run(ctx context.Context) error {
	b.mu.Lock()
	b.stopped = false
	b.running.Store(true)
	b.mu.Unlock()
	ticker := time.NewTicker(RescanInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			b.stop()
			return nil
		case <-ticker.C:
			b.Flush()
		}
	}
}

// Running reports whether the broadcaster is running, so it's watched, such
// as once Run, which is started in a goroutine, has started, and until it
// stops. A broadcaster which has stopped isn't watched until it's running
// again. One which has never run is watched, but isn't running.
func (b *Broadcaster) Running() bool {
	return b.running.Load()
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
// full misses it, and is owed a rescan of its table: a change of the table's
// schema and name, which is an update, without a row, so it reads the
// table again. A rescan is sent before the changes after it, once the
// channel can receive it, whatever the filter's action. The changes of a
// table which is owed a rescan aren't sent, as they're in it. A channel
// which needs every change should be buffered.
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
// it matches, which can receive it now, after the rescans it's owed, and
// returns how many couldn't, as their channels were full, so they missed
// it, and are owed a rescan of its table, or were owed one already. It never
// waits, so it can be called under a lock, which keeps the changes in the
// order they happened.
func (b *Broadcaster) Publish(change schema.Change) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	var missed int
	key := table{change.Schema, change.Table}
	for ch, w := range b.channels {
		// The rescans it's owed are sent first, whether or not the change
		// matches its filter, so a watcher is sent them as changes are
		// published. A change of a table it was owed a rescan of isn't sent,
		// as the rescan is after it, and while any rescans are still owed,
		// it's owed one of this change's table too, so a change doesn't
		// arrive before a rescan of a change before it.
		owed := w.owes[key]
		w.flush(ch)
		if !w.filter.Matches(change) {
			continue
		} else if owed {
			missed++
			continue
		} else if len(w.owed) > 0 {
			w.owe(key)
			missed++
			continue
		}
		select {
		case ch <- change:
		default:
			w.owe(key)
			missed++
		}
	}
	return missed
}

// Flush sends the rescans which watches are owed, which their channels can
// receive now. Run calls it at the RescanInterval, and Publish before each
// change, so it's only needed by a broadcaster which isn't run.
func (b *Broadcaster) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch, w := range b.channels {
		w.flush(ch)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// owe records that the watch is owed a rescan of the table
func (w *watch) owe(key table) {
	if w.owes == nil {
		w.owes = make(map[table]bool)
	}
	if !w.owes[key] {
		w.owes[key] = true
		w.owed = append(w.owed, key)
	}
}

// flush sends the rescans the watch is owed, in the order it missed them,
// which the channel can receive now, and stops at the first it can't
func (w *watch) flush(ch chan<- schema.Change) {
	for len(w.owed) > 0 {
		key := w.owed[0]
		select {
		case ch <- schema.Change{Action: schema.ChangeUpdate, Schema: key.schema, Table: key.name}:
			w.owed = w.owed[1:]
			delete(w.owes, key)
		default:
			return
		}
	}
}

// stop ends every watch, as the broadcaster has stopped running, and those
// made later, until it runs again
func (b *Broadcaster) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopped = true
	b.running.Store(false)
	for ch, w := range b.channels {
		close(w.stop)
		delete(b.channels, ch)
	}
}
