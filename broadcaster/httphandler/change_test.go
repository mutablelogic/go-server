package httphandler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	// Packages
	broadcaster "github.com/mutablelogic/go-server/broadcaster"
	httphandler "github.com/mutablelogic/go-server/broadcaster/httphandler"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	test "github.com/mutablelogic/go-server/broadcaster/test"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// streamed is an event of a stream, by its name, with its data
type streamed struct {
	name string
	data string
}

// newServer returns a test server of the broadcaster's API
func newServer(t *testing.T, b *broadcaster.Broadcaster) *test.Server {
	t.Helper()
	return test.NewServerFunc(t, func(router *httprouter.Router) error {
		return httphandler.RegisterHandlers(b, router)
	})
}

// watch requests the broadcaster's changes, with the accepted type, and returns
// the response, and the events which are streamed, which is closed when the
// stream ends
func watch(t *testing.T, server *test.Server, accept string) (*http.Response, <-chan streamed) {
	t.Helper()
	return watchQuery(t, server, accept, "")
}

// watchQuery requests the broadcaster's changes, as watch does, with the
// query, such as ?table=provider
func watchQuery(t *testing.T, server *test.Server, accept, query string) (*http.Response, <-chan streamed) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+test.Prefix+"/changes"+query, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", accept)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	events := make(chan streamed, 16)
	go func() {
		defer close(events)
		var name string
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if value, ok := strings.CutPrefix(line, "event: "); ok {
				name = value
			} else if value, ok := strings.CutPrefix(line, "data: "); ok && name != "" {
				events <- streamed{name, value}
				name = ""
			}
		}
	}()
	return resp, events
}

// next returns the next event of the stream, or fails when there's none
func next(t *testing.T, events <-chan streamed) streamed {
	t.Helper()
	select {
	case event, ok := <-events:
		require.True(t, ok, "the stream ended")
		return event
	case <-time.After(5 * time.Second):
		t.Fatal("no event")
		return streamed{}
	}
}

// decode returns the event of the stream
func decode(t *testing.T, event streamed) schema.Change {
	t.Helper()
	require.Equal(t, schema.ChangeName, event.name)
	var result schema.Change
	require.NoError(t, json.Unmarshal([]byte(event.data), &result))
	return result
}

func Test_WatchChanges(t *testing.T) {
	b := broadcaster.New()
	server := newServer(t, b)

	t.Run("Stream", func(t *testing.T) {
		// The events are streamed, from when the stream starts
		resp, events := watch(t, server, types.ContentTypeTextStream)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, types.ContentTypeTextStream, resp.Header.Get(types.ContentTypeHeader))
		created := schema.Change{Action: schema.ChangeCreate, Table: "provider", Row: "chat"}
		updated := schema.Change{Action: schema.ChangeUpdate, Table: "session", Row: "1", Data: json.RawMessage(`{"serial":2}`)}
		b.Publish(created)
		b.Publish(updated)
		assert.Equal(t, created, decode(t, next(t, events)))
		assert.Equal(t, updated, decode(t, next(t, events)))
	})

	t.Run("Any", func(t *testing.T) {
		// A client which accepts any type is streamed to
		resp, _ := watch(t, server, types.ContentTypeAny)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NotAcceptable", func(t *testing.T) {
		resp, _ := watch(t, server, types.ContentTypeJSON)
		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
	})

	t.Run("Filter", func(t *testing.T) {
		// Only the changes which match the query are streamed
		resp, events := watchQuery(t, server, types.ContentTypeTextStream, "?schema=llm&table=session&action=update")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		other := schema.Change{Action: schema.ChangeUpdate, Schema: "llm", Table: "provider", Row: "chat"}
		session := schema.Change{Action: schema.ChangeUpdate, Schema: "llm", Table: "session", Row: "2"}
		b.Publish(other)
		b.Publish(session)
		assert.Equal(t, session, decode(t, next(t, events)))
	})

	t.Run("BadFilter", func(t *testing.T) {
		resp, _ := watchQuery(t, server, types.ContentTypeTextStream, "?action=INSERT")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func Test_WatchChangesShutdown(t *testing.T) {
	// When the broadcaster stops running, as the server is stopping, a
	// shutdown event is sent, and the stream ends
	b := broadcaster.New()
	server := newServer(t, b)
	ctx, stop := context.WithCancel(context.Background())
	running := make(chan error)
	go func() { running <- b.Run(ctx) }()

	resp, events := watch(t, server, types.ContentTypeTextStream)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	stop()
	require.NoError(t, <-running)
	assert.Equal(t, schema.ChangeShutdown, next(t, events).name)
	select {
	case _, ok := <-events:
		assert.False(t, ok, "the stream didn't end")
	case <-time.After(5 * time.Second):
		t.Fatal("the stream didn't end")
	}

	// And it isn't streamed to again, until it runs again
	resp, _ = watch(t, server, types.ContentTypeTextStream)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.True(t, resp.Close, "the connection isn't closed")
}
