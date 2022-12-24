package tokenauth_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httpserver/router"
	tokenauth "github.com/mutablelogic/go-server/pkg/httpserver/tokenauth"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	plugin "github.com/mutablelogic/go-server/plugin"
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func body(w *httptest.ResponseRecorder) string {
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func Test_gateway_001(t *testing.T) {
	// Add a tokenauth gateway to a router. The prefixes for the services are
	// /api/v1/tokenauth and /api/v1/router
	assert := assert.New(t)
	plugins := []Plugin{
		tokenauth.WithLabel(t.Name()).WithPath(t.TempDir()),
		router.WithLabel("main").WithRoutes([]router.Route{
			{Prefix: "/api/v1/tokenauth", Handler: types.Task{Ref: "tokenauth." + t.Name()}},
		}).WithPrefix("/api/v1/router"),
	}
	provider, err := task.NewProvider(context.Background(), plugins...)
	assert.NoError(err)
	assert.NotNil(provider)

	router := provider.Get("router").(plugin.Router)
	assert.NotNil(router)

	t.Run("localhost/api/v1/router", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost/api/v1/router", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		t.Log(body(w))
	})

	t.Run("localhost/api/v1/router/api/v1/tokenauth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost/api/v1/router/api/v1/tokenauth", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		t.Log(body(w))
	})

	t.Run("localhost/api/v1/tokenauth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost/api/v1/tokenauth", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		t.Log(body(w))
	})

	t.Run("localhost/api/v1/tokenauth/admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost/api/v1/tokenauth/admin", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		t.Log(body(w))
	})

	t.Run("localhost/api/v1/tokenauth/not_admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "http://localhost/api/v1/tokenauth/not_admin", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusNotFound, w.Result().StatusCode)
		t.Log(body(w))
	})

	t.Run("localhost/api/v1/tokenauth/not_admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", "http://localhost/api/v1/tokenauth/admin", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusAccepted, w.Result().StatusCode)
		t.Log(body(w))
	})
}
