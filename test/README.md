# Test server

`test` is a HTTP test server for an API's unit tests.

- It serves the routes a function registers with a router, at `/api`.
- It records every request and response.
- When the test ends, it checks each one against the router's OpenAPI specification, and fails the test for any which doesn't conform.

So a test of a handler also checks that the handler does what its documentation says.

## What's checked

- **Operations:** the route which matched has the request's method documented.
- **Statuses:** the response's status is documented, or there's a `default` response.
- **Request bodies:** a successful request's body matches its schema.
- **Response bodies:** each response's body matches its schema, or is empty when none is documented.
- **Content types:** a body is matched by its content type, or a documented wildcard, such as `image/*` or `*/*`.
  - JSON bodies are validated directly.
  - YAML is validated as the JSON it's equivalent to.
  - A multipart form is validated as an object of its fields, with each file as its name.
  - A `text/plain` body, or a stream of server-sent events, is validated as a string.

## What's not checked

- **Redirects** from a path which isn't documented, as the request they redirect to is checked.
- **Unrouted paths:** a 404 for a path which no route matches.
- **Exchanges whose client went away**, such as a stream the test closed, or an interrupted upload, as their response can't be received.

## A test

```go
package mymodule_test

import (
	"net/http"
	"testing"

	"github.com/mutablelogic/go-server/pkg/httprouter"
	"github.com/mutablelogic/go-server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetThing(t *testing.T) {
	// Serve the module's API, which is checked when the test ends
	server := test.NewServer(t, func(router *httprouter.Router) error {
		return httphandler.RegisterHandlers(manager, router)
	})

	// GET /api/thing/one, and decode the response
	var thing schema.Thing
	require.Equal(t, http.StatusOK, server.JSON(t, http.MethodGet, "thing/one", nil, &thing))
	assert.Equal(t, "one", thing.Name)

	// A thing which doesn't exist isn't found
	assert.Equal(t, http.StatusNotFound, server.JSON(t, http.MethodGet, "thing/none", nil, nil))
}
```

Paths are relative to the prefix, and can have a query, such as `thing?limit=10`. A trailing slash is kept, as it's significant to the router.

## Requests

| Method | Sends | Returns |
|---|---|---|
| `JSON(t, method, path, body, v)` | a body as JSON, and decodes a 2xx response into `v` | the status |
| `Do(t, method, path, body)` | a body as JSON | the status and body |
| `DoWithContentType(t, method, path, contentType, body)` | a body with the content type | the status and body |
| `Send(t, method, path, header, body)` | the headers, and the body as it is | the status, headers and body |
| `Request(t, method, path, header)` | the headers, without a body | the status, headers and body |
| `Head(t, path, header)` | a `HEAD` request, to a route which handles it | the status and headers |

A body which is a `string` or `[]byte` is sent as it is, `nil` sends none, and any other value is encoded as JSON.

```go
// Create a thing, with a JSON body
var created schema.Thing
status := server.JSON(t, http.MethodPost, "thing", schema.ThingMeta{Name: "two"}, &created)
require.Equal(t, http.StatusCreated, status)

// Send YAML, which is checked against the request body's schema, when the
// operation documents application/yaml
status, body := server.DoWithContentType(t, http.MethodPost, "thing", "application/yaml", "name: three\n")

// Send headers, such as a precondition, and read the response's headers
status, header, body := server.Send(t, http.MethodGet, "thing/one", http.Header{
	"If-None-Match": {`"abc"`},
}, nil)
assert.Equal(t, http.StatusNotModified, status)
```

## A module's server

A module with its own handlers can wrap the server, so its tests make one call:

```go
// Package test has a test server of the module's API
package test

// NewServer starts a test server of the manager's API, which is checked
// against its OpenAPI specification when the test ends
func NewServer(t *testing.T, manager *mymodule.Manager) *test.Server {
	t.Helper()
	return test.NewServer(t, func(router *httprouter.Router) error {
		return errors.Join(
			httphandler.RegisterHandlers(manager, router),
			broadcasterhandler.RegisterHandlers(manager.Broadcaster(), router),
		)
	})
}
```

## Streams

A stream, such as of server-sent events, is requested with the server's URL and `test.Prefix`, so the test can read it as it arrives. Its exchange is recorded once the handler returns, and isn't checked when the test cancels the request.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+test.Prefix+"/changes", nil)
require.NoError(t, err)
req.Header.Set("Accept", "text/event-stream")
resp, err := http.DefaultClient.Do(req)
require.NoError(t, err)
defer resp.Body.Close()
```

A client of the API, such as one made with `go-client`, can be given `server.URL + test.Prefix` as its endpoint too.

## Testing what fails the check

`Verify` returns an error for each exchange which doesn't conform, without failing the test. `Reset` discards the exchanges, so a test which makes ones which don't conform, on purpose, doesn't fail when it ends.

```go
func TestUndocumented(t *testing.T) {
	server := test.NewServer(t, register)

	// A status which isn't documented doesn't conform
	server.Do(t, http.MethodGet, "thing?fail=teapot", nil)
	errs := server.Verify()
	require.Len(t, errs, 1)
	assert.ErrorContains(t, errs[0], "status 418 is not documented")

	// Discard it, so the test doesn't fail when it ends
	server.Reset()
}
```

`Exchanges` returns what's recorded: each request's method, path, the route pattern which matched it, headers and body, and its response's status, headers and body. `Spec` returns the OpenAPI specification which they're checked against.
