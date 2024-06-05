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
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_router_001(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(server.Router)
	router.AddHandlerFunc(context.Background(), "/hello", nil)
}

func Test_router_002(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)
	assert.NotNil(task)

	// Handle only the path /hello
	path := regexp.MustCompile("^/hello$")
	ctx := router.WithHost(context.Background(), "mutablelogic.com")
	task.(router.Router).AddHandlerFuncRe(ctx, path, func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("/hello", func(t *testing.T) {
		match, code := task.(router.Router).Match("mutablelogic.com", "GET", "/hello")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/", match.Prefix())
			assert.Equal("/hello", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})
	t.Run("/", func(t *testing.T) {
		_, code := task.(router.Router).Match("mutablelogic.com", "GET", "/")
		assert.Equal(404, code)
	})
}

func Test_router_003(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)
	assert.NotNil(task)

	// Handle all of mutablelogic.com and subdomains
	ctx := router.WithHost(context.Background(), "mutablelogic.com")
	task.(router.Router).AddHandlerFunc(ctx, "/", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("http://mutablelogic.com/", func(t *testing.T) {
		match, code := task.(router.Router).Match("mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/", match.Prefix())
			assert.Equal("/", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})

	t.Run("http://www.mutablelogic.com/", func(t *testing.T) {
		match, code := task.(router.Router).Match("www.mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/", match.Prefix())
			assert.Equal("/", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})

	t.Run("http://logic.com/", func(t *testing.T) {
		_, code := task.(router.Router).Match("logic.com", "GET", "/")
		assert.Equal(404, code)
	})

	t.Run("http://mutablelogic.com/hello", func(t *testing.T) {
		match, code := task.(router.Router).Match(".mutablelogic.com", "GET", "/hello")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/", match.Prefix())
			assert.Equal("/hello", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})
}

func Test_router_004(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)
	assert.NotNil(task)

	// Handle all of mutablelogic.com and subdomains with GET
	// We set an empty path which should be converted to /
	ctx := router.WithHost(context.Background(), "mutablelogic.com")
	task.(router.Router).AddHandlerFunc(ctx, "", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	}, "GET")

	t.Run("GET http://mutablelogic.com/", func(t *testing.T) {
		match, code := task.(router.Router).Match("www.mutablelogic.com", "GET", "/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/", match.Prefix())
			assert.Equal("/", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})

	t.Run("POST http://mutablelogic.com/", func(t *testing.T) {
		_, code := task.(router.Router).Match("mutablelogic.com", "POST", "/")
		assert.Equal(405, code)
	})
}

func Test_router_005(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)

	// Set a prefix as /api/v1
	ctx := router.WithHostPrefix(context.Background(), "mutablelogic.com", "/api/v1")
	task.(router.Router).AddHandlerFunc(ctx, "", func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})

	t.Run("GET http://mutablelogic.com/", func(t *testing.T) {
		_, code := task.(router.Router).Match("mutablelogic.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET http://mutablelogic.com/api/v1/", func(t *testing.T) {
		match, code := task.(router.Router).Match("mutablelogic.com", "GET", "/api/v1/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v1", match.Prefix())
			assert.Equal("/", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})
	t.Run("GET http://mutablelogic.com/api/v1/test", func(t *testing.T) {
		match, code := task.(router.Router).Match("mutablelogic.com", "GET", "/api/v1/test")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v1", match.Prefix())
			assert.Equal("/test", match.Path())
			assert.Equal(".mutablelogic.com", match.Host())
		}
	})
}

func Test_router_006(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)

	// Add two handlers which start with the same path
	ctx := router.WithPrefix(context.Background(), "/")
	task.(router.Router).AddHandlerFunc(provider.WithLabel(ctx, "first"), "/first", func(w http.ResponseWriter, r *http.Request) {})
	task.(router.Router).AddHandlerFunc(provider.WithLabel(ctx, "second"), "/first/second", func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := task.(router.Router).Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /first", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first", match.Path())
			assert.Equal("", match.Host())
		}
	})
	t.Run("GET /first/sec", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first/sec")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first/sec", match.Path())
			assert.Equal("", match.Host())
		}
	})
	t.Run("GET /first/second", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first/second")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first/second", match.Path())
			assert.Equal("", match.Host())
		}
	})
	t.Run("GET /first/second/third", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first/second/third")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first/second/third", match.Path())
			assert.Equal("", match.Host())
		}
	})
}

func Test_router_007(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)

	// Add two handlers which start with the same path
	ctx := router.WithPrefix(context.Background(), "/")
	task.(router.Router).AddHandlerFuncRe(provider.WithLabel(ctx, "first"), regexp.MustCompile("^/first(.*)$"), func(w http.ResponseWriter, r *http.Request) {})
	task.(router.Router).AddHandlerFuncRe(provider.WithLabel(ctx, "second"), regexp.MustCompile("second(.*)$"), func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := task.(router.Router).Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /first", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first", match.Path())
			assert.Equal("", match.Host())
			assert.Equal([]string{""}, match.Parameters())
		}
	})
	t.Run("GET /first/sec", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first/sec")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first/sec", match.Path())
			assert.Equal("", match.Host())
			assert.Equal([]string{"/sec"}, match.Parameters())
		}
	})
	t.Run("GET /first/second", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/first/second")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("first", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/first/second", match.Path())
			assert.Equal("", match.Host())
			assert.Equal([]string{"/second"}, match.Parameters())
		}
	})
	t.Run("GET /first/second/third", func(t *testing.T) {
		match, code := task.(router.Router).Match("any.com", "GET", "/second/third")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("second", match.Label())
			assert.Equal("/", match.Prefix())
			assert.Equal("/second/third", match.Path())
			assert.Equal("", match.Host())
			assert.Equal([]string{"/third"}, match.Parameters())
		}
	})
}

func Test_router_008(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)

	// Add one handler for all paths
	ctx := router.WithPrefix(context.Background(), "/api/v2/")
	task.(router.Router).AddHandlerFuncRe(ctx, regexp.MustCompile("^/(.*)$"), func(w http.ResponseWriter, r *http.Request) {})

	t.Run("GET /", func(t *testing.T) {
		_, code := task.(router.Router).Match("any.com", "GET", "/")
		assert.Equal(404, code)
	})
	t.Run("GET /api/v2", func(t *testing.T) {
		match, code := task.(router.Router).Match("", "GET", "/api/v2")
		assert.Equal(308, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v2/", match.Path())
		}
	})
	t.Run("GET /api/v2/", func(t *testing.T) {
		match, code := task.(router.Router).Match("", "GET", "/api/v2/")
		assert.Equal(200, code)
		if assert.NotNil(match) {
			assert.Equal("/api/v2", match.Prefix())
			assert.Equal("/", match.Path())
			assert.Equal("", match.Host())
			assert.Equal([]string{""}, match.Parameters())
		}
	})

}

func Test_router_009(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New()
	assert.NoError(err)

	ctx := router.WithPrefix(context.Background(), "/api/v2")
	task.(router.Router).AddHandlerFuncRe(ctx, regexp.MustCompile("^/(.*)$"), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	server := httptest.NewServer(task.(http.Handler))
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
