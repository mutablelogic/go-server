package httphandler

import (
	"context"
	"errors"
	"net/http"
	"time"

	// Packages
	broadcaster "github.com/mutablelogic/go-server/broadcaster"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// changeBuffer is how many changes are kept for a client, while it reads
// the ones before them, before they're dropped
const changeBuffer = 16

// changeWriteTimeout is the longest a write to a client's stream can take,
// before the client is taken to have stopped reading
var changeWriteTimeout = 30 * time.Second

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func changeHandlers(b *broadcaster.Broadcaster, router *httprouter.Router, documentation *openapi.MarkdownDoc) error {
	router.Spec().AddTag(Tag, documentation.Section(2, "Change Operations").Body)
	return router.Register("changes", nil, func(path httprequest.PathItem) {
		path.Tag(Tag)

		// GET
		path.Get(changeWatchHandler(b), func(op httprequest.PathOperation) {
			op.Summary("Watch Changes")
			op.Description(documentation.Section(3, "GET /changes").Body)
			op.Query(jsonschema.MustFor[schema.ChangeFilter]())
			op.TextStreamResponse(http.StatusOK)
			op.ErrorResponse(http.StatusBadRequest, "Invalid request, such as an unknown action")
			op.ErrorResponse(http.StatusNotAcceptable, "The response can't be any of the accepted types")
			op.ErrorResponse(http.StatusServiceUnavailable, "The server is stopping")
		})
	})
}

// changeWatchHandler streams the broadcaster's changes, as server-sent
// events, until the client goes away, or the broadcaster stops, as the
// server is stopping, when a shutdown event is sent before the stream ends
func changeWatchHandler(b *broadcaster.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Changes are only streamed
		if accept, err := types.AcceptContentType(r); err != nil {
			httpresponse.Error(w, httpresponse.ErrBadRequest.With(err.Error()))
			return
		} else if accept != "" && accept != types.ContentTypeAny && accept != types.ContentTypeTextStream {
			httpresponse.Error(w, httpresponse.Err(http.StatusNotAcceptable).Withf("cannot respond with %q", accept))
			return
		}

		// The changes which are streamed
		var filter schema.ChangeFilter
		if err := httprequest.Query(r.URL.Query(), &filter); err != nil {
			httpresponse.Error(w, badRequest(err))
			return
		}

		// Watch the broadcaster, until the client goes away, before the stream
		// starts, so a client which loads what it shows once the stream has
		// started misses no changes
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		ch := make(chan schema.Change, changeBuffer)

		// The server is stopping when the broadcaster has stopped, so the
		// client isn't streamed to, and its connection is closed, so the
		// server isn't kept waiting for it
		done, err := b.Watch(ctx, filter, ch)
		if err != nil {
			if errors.Is(err, httpresponse.ErrServiceUnavailable) {
				w.Header().Set("Connection", "close")
			}
			httpresponse.Error(w, err)
			return
		}

		// The stream lasts as long as the client watches, so it isn't cut off
		// by the server's write timeout, which is for the whole response, but
		// each write has its own deadline. A proxy, such as nginx, doesn't
		// buffer it. A writer which can't have its deadline set is streamed
		// to until the server's timeout.
		dw, err := httpresponse.NewDeadlineWriter(w, changeWriteTimeout)
		if err != nil {
			httpresponse.Error(w, httpresponse.ErrInternalError.With(err.Error()))
			return
		}

		// Stream the changes
		stream := httpresponse.NewTextStream(dw, "X-Accel-Buffering", "no")
		if stream == nil {
			httpresponse.Error(w, httpresponse.ErrInternalError.With("streaming isn't supported"))
			return
		}
		defer stream.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case <-dw.Failed():
				// The client stopped reading, or went away
				return
			case err := <-done:
				// The broadcaster stopped, as the server is stopping, so the
				// client is told, and the stream ends, so the server isn't
				// waiting for it. The event has data, as an event without
				// any isn't dispatched by a browser.
				if err != nil {
					stream.Write(schema.ChangeShutdown, struct{}{})
				}
				return
			case change := <-ch:
				stream.Write(schema.ChangeName, change)
			}
		}
	}
}

// badRequest returns the error as a bad request, unless it has its status
// already, such as an error of a query
func badRequest(err error) error {
	var status httpresponse.Err
	if errors.As(err, &status) {
		return err
	}
	return httpresponse.ErrBadRequest.With(err.Error())
}
