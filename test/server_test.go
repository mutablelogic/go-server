package test_test

import (
	"net/http"
	"testing"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	test "github.com/mutablelogic/go-server/test"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// thing is the response of the test API
type thing struct {
	Name string `json:"name"`
}

// newServer returns a test server of an API with one route, thing, which
// returns a thing, or, with ?status=, that status, which isn't documented,
// or, with ?bad=, a body which doesn't match the schema
func newServer(t *testing.T) *test.Server {
	t.Helper()
	return test.NewServer(t, func(router *httprouter.Router) error {
		return router.Register("thing", nil, func(path httprequest.PathItem) {
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Query().Has("status"):
					httpresponse.Error(w, httpresponse.Err(http.StatusTeapot))
				case r.URL.Query().Has("bad"):
					httpresponse.JSON(w, http.StatusOK, 0, map[string]int{"name": 1})
				default:
					httpresponse.JSON(w, http.StatusOK, 0, thing{Name: "one"})
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Get a thing")
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[thing]())
			})
		})
	})
}

func Test_Server(t *testing.T) {
	t.Run("Conforms", func(t *testing.T) {
		// A response which the specification documents conforms, and the
		// requests' helpers send to the API's prefix
		server := newServer(t)
		var result thing
		require.Equal(t, http.StatusOK, server.JSON(t, http.MethodGet, "thing", nil, &result))
		assert.Equal(t, "one", result.Name)
		status, header, body := server.Send(t, http.MethodGet, "thing", http.Header{"Accept": {"application/json"}}, nil)
		assert.Equal(t, http.StatusOK, status)
		assert.Contains(t, header.Get("Content-Type"), "application/json")
		assert.JSONEq(t, `{"name":"one"}`, string(body))
		require.Len(t, server.Exchanges(), 2)
		assert.Equal(t, "/api/thing", server.Exchanges()[0].Path)
		assert.Empty(t, server.Verify())
	})

	t.Run("Undocumented", func(t *testing.T) {
		// A status which isn't documented, a body which doesn't match the
		// schema, and a path which isn't routed, aren't conforming, except
		// that a path which isn't routed isn't found
		server := newServer(t)
		status, _, _ := server.Request(t, http.MethodGet, "thing?status=1", nil)
		assert.Equal(t, http.StatusTeapot, status)
		server.Do(t, http.MethodGet, "thing?bad=1", nil)
		status, _ = server.Do(t, http.MethodGet, "nothing", nil)
		assert.Equal(t, http.StatusNotFound, status)
		errs := server.Verify()
		require.Len(t, errs, 2)
		assert.ErrorContains(t, errs[0], "status 418 is not documented")
		assert.ErrorContains(t, errs[1], "response body")

		// The exchanges are reset, so the test doesn't fail when it ends
		server.Reset()
		assert.Empty(t, server.Exchanges())
	})
}
