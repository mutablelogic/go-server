package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-server/broadcaster/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WatchChanges calls the function with each of the broadcaster's changes
// which match the filter, which are all of them when it's empty, as the
// server streams them, such as when a row is created, updated or deleted,
// until the context is cancelled, when it returns nil, or the
// function returns an error, which it returns, or the stream ends. When the
// server stops, it returns ErrServiceUnavailable.
func (c *Client) WatchChanges(ctx context.Context, filter schema.ChangeFilter, fn func(schema.Change) error) error {
	if fn == nil {
		return httpresponse.ErrBadRequest.With("missing function")
	}

	// Read the changes, and ignore the pings. The function's error is kept,
	// as it's returned even when the context is cancelled as well.
	var fnErr, shutdown error
	callback := func(event client.TextStreamEvent) error {
		if event.Event == schema.ChangeShutdown {
			shutdown = httpresponse.ErrServiceUnavailable.With("server is stopping")
			return nil
		} else if event.Event != schema.ChangeName {
			return nil
		}
		var data schema.Change
		if err := event.Json(&data); err != nil {
			return fmt.Errorf("malformed %s event: %w", event.Event, err)
		}
		fnErr = fn(data)
		return fnErr
	}
	err := c.DoWithContext(ctx, client.MethodGet, nil,
		client.OptPath("changes"),
		client.OptQuery(filter.Query()),
		client.OptReqHeader("Accept", types.ContentTypeTextStream),
		client.OptTextStreamCallback(callback),
		client.OptNoTimeout(),
	)
	switch {
	case fnErr != nil:
		return fnErr
	case ctx.Err() != nil:
		return nil
	case shutdown != nil:
		return shutdown
	case isUnavailable(err):
		// The server was stopping before the stream started, so it's
		// ErrServiceUnavailable, as when it stops while streaming, with the
		// response's error
		return errors.Join(httpresponse.ErrServiceUnavailable.With("server is stopping"), err)
	default:
		return err
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// isUnavailable returns true when the error is a response whose status is
// 503 Service Unavailable
func isUnavailable(err error) bool {
	var response httpresponse.ErrResponse
	return errors.As(err, &response) && response.Code == http.StatusServiceUnavailable
}
