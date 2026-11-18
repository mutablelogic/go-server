package broadcaster_test

import (
	"context"
	"testing"
	"time"

	// Packages
	broadcaster "github.com/mutablelogic/go-server/broadcaster"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// watch watches the broadcaster with a buffered channel, until the test
// ends, and returns the channel
func watch(t *testing.T, b *broadcaster.Broadcaster) chan schema.Change {
	t.Helper()
	ch := make(chan schema.Change, 16)
	ctx, cancel := context.WithCancel(context.Background())
	done, err := b.Watch(ctx, schema.ChangeFilter{}, ch)
	require.NoError(t, err)
	t.Cleanup(func() {
		cancel()
		assert.NoError(t, <-done)
	})
	return ch
}

// events returns the events which have been sent to the channel
func events(ch chan schema.Change) []schema.Change {
	var result []schema.Change
	for {
		select {
		case event := <-ch:
			result = append(result, event)
		default:
			return result
		}
	}
}

func Test_Broadcaster(t *testing.T) {
	ctx := context.Background()
	created := schema.Change{Action: schema.ChangeCreate, Table: "provider", Row: "ollama"}
	deleted := schema.Change{Action: schema.ChangeDelete, Table: "provider", Row: "ollama"}

	t.Run("Publish", func(t *testing.T) {
		// Each watcher is sent the events, in the order they're published
		b := broadcaster.New()
		a, c := watch(t, b), watch(t, b)
		b.Publish(created)
		b.Publish(deleted)
		assert.Equal(t, []schema.Change{created, deleted}, events(a))
		assert.Equal(t, []schema.Change{created, deleted}, events(c))
	})

	t.Run("Filter", func(t *testing.T) {
		// A watcher is sent the changes which match its filter, and
		// Publish returns how many watchers missed a change, as their
		// channels were full
		b := broadcaster.New()
		all := watch(t, b)
		providers := make(chan schema.Change, 1)
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := b.Watch(watchCtx, schema.ChangeFilter{Table: "provider", Action: schema.ChangeDelete}, providers)
		require.NoError(t, err)

		assert.Zero(t, b.Publish(created))
		assert.Zero(t, b.Publish(deleted))
		assert.Equal(t, []schema.Change{created, deleted}, events(all))
		assert.Equal(t, []schema.Change{deleted}, events(providers))

		// The filtered channel holds one change, so the second is missed
		assert.Zero(t, b.Publish(deleted))
		assert.Equal(t, 1, b.Publish(deleted))
		events(all)
		events(providers)

		// An unknown action isn't a filter
		_, err = b.Watch(watchCtx, schema.ChangeFilter{Action: "INSERT"}, make(chan schema.Change))
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)
	})

	t.Run("Errors", func(t *testing.T) {
		b := broadcaster.New()
		_, err := b.Watch(ctx, schema.ChangeFilter{}, nil)
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		ch := make(chan schema.Change, 1)
		_, err = b.Watch(watchCtx, schema.ChangeFilter{}, ch)
		require.NoError(t, err)
		_, err = b.Watch(watchCtx, schema.ChangeFilter{}, ch)
		assert.ErrorIs(t, err, httpresponse.ErrConflict)
	})

	t.Run("Cancelled", func(t *testing.T) {
		// Once the watch ends, no more events are sent, so the channel can
		// be closed, and it can watch again
		b := broadcaster.New()
		ch := make(chan schema.Change, 16)
		watchCtx, cancel := context.WithCancel(ctx)
		done, err := b.Watch(watchCtx, schema.ChangeFilter{}, ch)
		require.NoError(t, err)
		cancel()
		require.NoError(t, <-done)
		close(ch)
		b.Publish(created)

		ch = watch(t, b)
		b.Publish(deleted)
		assert.Equal(t, []schema.Change{deleted}, events(ch))
	})

	t.Run("Stopped", func(t *testing.T) {
		// When the broadcaster stops running, the watches end, with
		// ErrServiceUnavailable, and it isn't watched until it runs again
		b := broadcaster.New()
		runCtx, stop := context.WithCancel(ctx)
		running := make(chan error)
		go func() { running <- b.Run(runCtx) }()
		ch := make(chan schema.Change, 16)
		done, err := b.Watch(ctx, schema.ChangeFilter{}, ch)
		require.NoError(t, err)
		stop()
		require.NoError(t, <-running)
		assert.ErrorIs(t, <-done, httpresponse.ErrServiceUnavailable)
		_, err = b.Watch(ctx, schema.ChangeFilter{}, ch)
		assert.ErrorIs(t, err, httpresponse.ErrServiceUnavailable)

		// And it runs again, so it's watched once it's running, until it's
		// stopped, which is waited for, so it doesn't outlast the test
		runCtx, stop = context.WithCancel(ctx)
		go func() { running <- b.Run(runCtx) }()
		require.Eventually(t, b.Running, time.Second, time.Millisecond)
		done, err = b.Watch(ctx, schema.ChangeFilter{}, make(chan schema.Change))
		require.NoError(t, err)
		stop()
		require.NoError(t, <-running)
		assert.ErrorIs(t, <-done, httpresponse.ErrServiceUnavailable)
	})

	t.Run("Running", func(t *testing.T) {
		// A broadcaster which has never run is watched, but isn't running,
		// and it's running while it runs, and not once it has stopped
		b := broadcaster.New()
		assert.False(t, b.Running())
		_, err := b.Watch(ctx, schema.ChangeFilter{}, make(chan schema.Change))
		require.NoError(t, err)
		runCtx, stop := context.WithCancel(ctx)
		running := make(chan error)
		go func() { running <- b.Run(runCtx) }()
		require.Eventually(t, b.Running, time.Second, time.Millisecond)
		stop()
		require.NoError(t, <-running)
		assert.False(t, b.Running())
	})

	t.Run("Full", func(t *testing.T) {
		// A channel which is full misses events, rather than the publisher
		// waiting for it
		b := broadcaster.New()
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := b.Watch(watchCtx, schema.ChangeFilter{}, make(chan schema.Change))
		require.NoError(t, err)
		published := make(chan struct{})
		go func() { b.Publish(created); close(published) }()
		select {
		case <-published:
		case <-time.After(time.Second):
			t.Fatal("publishing waited for a watcher")
		}
	})

	t.Run("Rescan", func(t *testing.T) {
		// A watcher which misses a change is owed a rescan of its table: an
		// update of it, without a row, which is sent once it can receive
		// it, before the changes after it, which aren't sent of the table,
		// as they're in it
		b := broadcaster.New()
		ch := make(chan schema.Change, 1)
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := b.Watch(watchCtx, schema.ChangeFilter{}, ch)
		require.NoError(t, err)
		other := schema.Change{Action: schema.ChangeCreate, Schema: "llm", Table: "session", Row: "1"}
		rescan := schema.Change{Action: schema.ChangeUpdate, Table: "provider"}
		assert.Zero(t, b.Publish(created))
		assert.Equal(t, 1, b.Publish(deleted))
		assert.Equal(t, 1, b.Publish(deleted), "it's owed a rescan already")
		assert.Equal(t, []schema.Change{created}, events(ch))

		// The rescan is sent first, so a change of another table, after it,
		// waits too, and is owed a rescan of its own
		assert.Equal(t, 1, b.Publish(other))
		assert.Equal(t, []schema.Change{rescan}, events(ch))
		b.Flush()
		assert.Equal(t, []schema.Change{{Action: schema.ChangeUpdate, Schema: "llm", Table: "session"}}, events(ch))

		// Once it's caught up, it's sent the changes again
		assert.Zero(t, b.Publish(created))
		assert.Equal(t, []schema.Change{created}, events(ch))

		// A rescan is sent whatever the filter's action, and Run sends the
		// rescans which are owed, without a change being published
		creates := make(chan schema.Change, 1)
		_, err = b.Watch(watchCtx, schema.ChangeFilter{Action: schema.ChangeCreate}, creates)
		require.NoError(t, err)
		assert.Zero(t, b.Publish(created))
		assert.Equal(t, 2, b.Publish(created), "both are full")
		events(ch)
		<-creates
		runCtx, stop := context.WithCancel(ctx)
		defer stop()
		go func() { _ = b.Run(runCtx) }()
		select {
		case change := <-creates:
			assert.Equal(t, rescan, change)
		case <-time.After(time.Second):
			t.Fatal("waiting for the rescan, which Run sends")
		}
	})

	t.Run("RescanOwed", func(t *testing.T) {
		b := broadcaster.New()
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		rescan := schema.Change{Action: schema.ChangeUpdate, Table: "provider"}
		other := schema.Change{Action: schema.ChangeCreate, Schema: "llm", Table: "session", Row: "1"}

		// A change of a table it's owed a rescan of isn't sent after the
		// rescan, though there's room for both, as the rescan is after it
		ch := make(chan schema.Change, 2)
		_, err := b.Watch(watchCtx, schema.ChangeFilter{}, ch)
		require.NoError(t, err)
		b.Publish(created)
		b.Publish(created)
		assert.Equal(t, 1, b.Publish(created))
		events(ch)
		assert.Equal(t, 1, b.Publish(deleted), "it's in the rescan, which is sent now")
		assert.Equal(t, []schema.Change{rescan}, events(ch))

		// A watcher is sent the rescan it's owed when a change which doesn't
		// match its filter is published
		providers := make(chan schema.Change, 1)
		_, err = b.Watch(watchCtx, schema.ChangeFilter{Table: "provider"}, providers)
		require.NoError(t, err)
		b.Publish(created)
		b.Publish(created)
		assert.Equal(t, 2, b.Publish(created), "both are full")
		events(ch)
		events(providers)
		b.Publish(other)
		assert.Equal(t, []schema.Change{rescan}, events(providers), "of another table, which it doesn't watch")
	})
}
