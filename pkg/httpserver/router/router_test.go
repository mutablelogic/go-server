package router_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	// Module import
	router "github.com/mutablelogic/go-server/pkg/httpserver/router"
	util "github.com/mutablelogic/go-server/pkg/httpserver/util"
	task "github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////
// TESTS

type response struct {
	Test   string
	Prefix string
	Path   string
	Params []string
}

func body(w *httptest.ResponseRecorder) string {
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func Test_router_001(t *testing.T) {
	assert := assert.New(t)
	plugin := router.WithLabel("main")
	provider, err := task.NewProvider(context.Background(), plugin)
	assert.NoError(err)
	router, ok := provider.Get(plugin.Name(), plugin.Label()).(router.Router)
	assert.True(ok)
	assert.NotNil(router)

	// Add request handlers: The add handler order is important
	router.AddHandlerEx("/test", nil, func(w http.ResponseWriter, req *http.Request) {
		prefix, path, params := util.ReqPrefixPathParams(req)
		util.ServeJSON(w, response{"1", prefix, path, params}, http.StatusOK, 0)
	})
	router.AddHandlerEx("/test", regexp.MustCompile(`^/(\d+)`), func(w http.ResponseWriter, req *http.Request) {
		prefix, path, params := util.ReqPrefixPathParams(req)
		util.ServeJSON(w, response{"3", prefix, path, params}, http.StatusOK, 0)
	})
	router.AddHandlerEx("/test", regexp.MustCompile(`^/(\d+)$`), func(w http.ResponseWriter, req *http.Request) {
		prefix, path, params := util.ReqPrefixPathParams(req)
		util.ServeJSON(w, response{"2", prefix, path, params}, http.StatusOK, 0)
	})
	router.AddHandlerEx("otherhost/test", regexp.MustCompile(`^/(\d+)`), func(w http.ResponseWriter, req *http.Request) {
		prefix, path, params := util.ReqPrefixPathParams(req)
		util.ServeJSON(w, response{"4", prefix, path, params}, http.StatusOK, 0)
	})

	t.Run("localhost/", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match 404
		req, err := http.NewRequest("GET", "http://localhost/", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusNotFound, w.Result().StatusCode)
	})
	t.Run("localhost/test", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match request #1
		req, err := http.NewRequest("GET", "http://localhost/test", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		assert.Equal(`{"Test":"1","Prefix":"/test","Path":"/","Params":null}`+"\n", body(w))
	})
	t.Run("localhost/test/", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match request #1
		req, err := http.NewRequest("GET", "http://localhost/test/", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		assert.Equal(`{"Test":"1","Prefix":"/test","Path":"/","Params":null}`+"\n", body(w))
	})
	t.Run("localhost/test99", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should return 404 error
		req, err := http.NewRequest("GET", "http://localhost/test99", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusNotFound, w.Result().StatusCode)
	})
	t.Run("localhost/test/99", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match request #2
		req, err := http.NewRequest("GET", "http://localhost/test/99", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		assert.Equal(`{"Test":"2","Prefix":"/test","Path":"/99","Params":["99"]}`+"\n", body(w))
	})
	t.Run("localhost/test/90/99", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match request #3
		req, err := http.NewRequest("GET", "http://localhost/test/90/99", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		assert.Equal(`{"Test":"3","Prefix":"/test","Path":"/90/99","Params":["90"]}`+"\n", body(w))
	})
	t.Run("www.otherhost/test/99/98", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should match request #4
		req, err := http.NewRequest("GET", "http://www.otherhost/test/99/98", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Result().StatusCode)
		assert.Equal(`{"Test":"4","Prefix":"/test","Path":"/99/98","Params":["99"]}`+"\n", body(w))
	})
	t.Run("www.otherhost", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should not match any request
		req, err := http.NewRequest("GET", "http://www.otherhost/", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusNotFound, w.Result().StatusCode)
	})
	t.Run("www.otherhost/test/99/98", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Should return MethodNotAllowed error
		req, err := http.NewRequest("POST", "http://www.otherhost/test/99/98", nil)
		assert.NoError(err)
		assert.NotNil(req)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusMethodNotAllowed, w.Result().StatusCode)
	})

}

func Test_router_002(t *testing.T) {
	// Create router which registers itself as a gateway
	plugin := router.WithLabel("main").WithRoutes([]router.Route{
		{
			Prefix:  "/api/v1/health",
			Handler: types.Task{Ref: "router.main"},
		},
	})

	// Create a provider, register router
	provider, err := task.NewProvider(context.Background(), plugin)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}

}
