package router_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/stretchr/testify/assert"
)

func Test_router_001(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(server.Router)
	router.AddHandler(context.Background(), "/hello", nil)
}

func Test_router_002(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(router.Router)

	// Handle only the path /hello
	path := regexp.MustCompile("^/hello$")
	router.AddHandlerRe(context.Background(), "mutablelogic.com", path, func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("/hello", func(t *testing.T) {
		match, code := router.Match("mutablelogic.com", "GET", "/hello")
		assert.Equal(200, code)
		assert.Equal(match.Prefix, "/")
		assert.Equal(match.Path, "/hello")
		assert.Equal(match.Host, ".mutablelogic.com")
	})
	t.Run("/", func(t *testing.T) {
		_, code := router.Match("mutablelogic.com", "GET", "/")
		assert.Equal(404, code)
	})
}

func Test_router_003(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(router.Router)

	// Handle all of mutablelogic.com and subdomains
	router.AddHandler(context.Background(), "mutablelogic.com", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("http://mutablelogic.com/", func(t *testing.T) {
		match, code := router.Match("mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		assert.Equal(match.Prefix, "/")
		assert.Equal(match.Path, "/")
		assert.Equal(match.Host, ".mutablelogic.com")
	})

	t.Run("http://www.mutablelogic.com/", func(t *testing.T) {
		match, code := router.Match("www.mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		assert.Equal(match.Prefix, "/")
		assert.Equal(match.Path, "/")
		assert.Equal(match.Host, ".mutablelogic.com")
	})

	t.Run("http://logic.com/", func(t *testing.T) {
		_, code := router.Match("logic.com", "GET", "/")
		assert.Equal(404, code)
	})

	t.Run("http://mutablelogic.com/hello", func(t *testing.T) {
		match, code := router.Match(".mutablelogic.com", "GET", "/hello")
		assert.Equal(200, code)
		assert.Equal("/", match.Prefix)
		assert.Equal("/hello", match.Path)
		assert.Equal(".mutablelogic.com", match.Host)
	})
}

func Test_router_004(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(router.Router)

	// Handle all of mutablelogic.com and subdomains with GET
	router.AddHandler(context.Background(), "mutablelogic.com", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}, "GET")

	t.Run("GET http://mutablelogic.com/", func(t *testing.T) {
		match, code := router.Match("mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		assert.Equal("/", match.Prefix)
		assert.Equal("/", match.Path)
		assert.Equal(".mutablelogic.com", match.Host)
	})

	t.Run("POST http://mutablelogic.com/", func(t *testing.T) {
		_, code := router.Match("mutablelogic.com", "POST", "/")
		assert.Equal(405, code)
	})
}

func Test_router_005(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)

	// Set a prefix as /api/v1
	ctx := router.WithPrefix(context.Background(), "/api/v1")
	router := task.(router.Router)
	router.AddHandler(ctx, "mutablelogic.com", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("GET http://mutablelogic.com/", func(t *testing.T) {
		_, code := router.Match("mutablelogic.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET http://mutablelogic.com/api/v1/", func(t *testing.T) {
		match, code := router.Match("mutablelogic.com", "GET", "/api/v1/")
		assert.Equal(200, code)
		assert.Equal("/api/v1", match.Prefix)
		assert.Equal("/", match.Path)
		assert.Equal(".mutablelogic.com", match.Host)
	})
	t.Run("GET http://mutablelogic.com/api/v1/test", func(t *testing.T) {
		match, code := router.Match("mutablelogic.com", "GET", "/api/v1/test")
		assert.Equal(200, code)
		assert.Equal("/api/v1", match.Prefix)
		assert.Equal("/test", match.Path)
		assert.Equal(".mutablelogic.com", match.Host)
	})
}

func Test_router_006(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)

	// Add two handlers which start with the same path
	ctx := router.WithPrefix(context.Background(), "/")
	r := task.(router.Router)
	r.AddHandler(router.WithKey(ctx, "first"), "/first", func(w http.ResponseWriter, r *http.Request) {})
	r.AddHandler(router.WithKey(ctx, "second"), "/first/second", func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := r.Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /first", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first", match.Path)
			assert.Equal("", match.Host)
		}
	})
	t.Run("GET /first/sec", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first/sec")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first/sec", match.Path)
			assert.Equal("", match.Host)
		}
	})
	t.Run("GET /first/second", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first/second")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first/second", match.Path)
			assert.Equal("", match.Host)
		}
	})
	t.Run("GET /first/second/third", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first/second/third")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first/second/third", match.Path)
			assert.Equal("", match.Host)
		}
	})
}

func Test_router_007(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)

	// Add two handlers which start with the same path
	ctx := router.WithPrefix(context.Background(), "/")
	r := task.(router.Router)
	r.AddHandlerRe(router.WithKey(ctx, "first"), "", regexp.MustCompile("^/first(.*)$"), func(w http.ResponseWriter, r *http.Request) {})
	r.AddHandlerRe(router.WithKey(ctx, "second"), "", regexp.MustCompile("second(.*)$"), func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := r.Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /first", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first", match.Path)
			assert.Equal("", match.Host)
			assert.Equal([]string{""}, match.Parameters)
		}
	})
	t.Run("GET /first/sec", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first/sec")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first/sec", match.Path)
			assert.Equal("", match.Host)
			assert.Equal([]string{"/sec"}, match.Parameters)
		}
	})
	t.Run("GET /first/second", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/first/second")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/first/second", match.Path)
			assert.Equal("", match.Host)
			assert.Equal([]string{"/second"}, match.Parameters)
		}
	})
	t.Run("GET /first/second/third", func(t *testing.T) {
		match, code := r.Match("any.com", "GET", "/second/third")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Key)
			assert.Equal("/", match.Prefix)
			assert.Equal("/second/third", match.Path)
			assert.Equal("", match.Host)
			assert.Equal([]string{"/third"}, match.Parameters)
		}
	})
}

func Test_router_008(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)

	// Add one handler for all paths
	ctx := router.WithPrefix(context.Background(), "/api/v2/")
	r := task.(router.Router)
	r.AddHandlerRe(ctx, "", regexp.MustCompile("^/(.*)$"), func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := r.Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /api/v2", func(t *testing.T) {
		match, code := r.Match("", "GET", "/api/v2")
		assert.Equal(308, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v2/", match.Path)
		}
	})
	t.Run("GET /api/v2/", func(t *testing.T) {
		match, code := r.Match("", "GET", "/api/v2/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v2", match.Prefix)
			assert.Equal("/", match.Path)
			assert.Equal("", match.Host)
			assert.Equal([]string{""}, match.Parameters)
		}
	})

}

func Test_router_009(t *testing.T) {
	assert := assert.New(t)
	handler, err := router.Config{}.New(context.Background())
	assert.NoError(err)

	ctx := router.WithPrefix(context.Background(), "/api/v2")
	handler.(server.Router).AddHandlerRe(ctx, "", regexp.MustCompile("^/(.*)$"), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	server := httptest.NewServer(handler.(http.Handler))
	defer server.Close()

	t.Run("GET /api/v2/", func(t *testing.T) {
		client := server.Client()
		res, err := client.Get(server.URL + "/api/v2")
		if assert.NoError(err) {
			assert.Equal(200, res.StatusCode)
			greeting, _ := io.ReadAll(res.Body)
			res.Body.Close()
			assert.Equal("Hello, World!", string(greeting))
		}
	})
}
