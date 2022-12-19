package router_test

import (
	"net/http"
	"regexp"
	"testing"

	// Module import
	router "github.com/mutablelogic/go-server/pkg/httpserver/router"
	assert "github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_route_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("*", func(t *testing.T) {
		route := router.NewRoute("/test", nil, func(w http.ResponseWriter, req *http.Request) {})
		t.Log(route)
		assert.NotNil(route)
		assert.Equal("*", route.Host())
		assert.Equal("/test", route.Path())
		assert.Equal([]string{"GET"}, route.Methods())
		assert.True(route.MatchesHost("localhost"))
		_, path, matches := route.MatchesPath("/test/test")
		assert.True(matches)
		assert.Equal(path, "/test")
		assert.True(route.MatchesMethod("GET"))
	})
	t.Run("host.com", func(t *testing.T) {
		route := router.NewRoute("host.com/test", nil, func(w http.ResponseWriter, req *http.Request) {})
		t.Log(route)
		assert.NotNil(route)
		assert.Equal("*.host.com", route.Host())
		assert.Equal("/test", route.Path())
		assert.Equal([]string{"GET"}, route.Methods())
		assert.False(route.MatchesHost("localhost"))
		_, path, matches := route.MatchesPath("/test/test")
		assert.True(matches)
		assert.Equal(path, "/test")
		assert.True(route.MatchesMethod("GET"))
	})
	t.Run("host.com", func(t *testing.T) {
		route := router.NewRoute("host.com", nil, func(w http.ResponseWriter, req *http.Request) {})
		t.Log(route)
		assert.NotNil(route)
		assert.Equal("*.host.com", route.Host())
		assert.Equal("/", route.Path())
		assert.Equal([]string{"GET"}, route.Methods())
		assert.False(route.MatchesHost("localhost"))
		assert.True(route.MatchesHost("www.host.com."))
		assert.True(route.MatchesHost("www.host.com"))
		assert.True(route.MatchesHost("host.com."))
		assert.True(route.MatchesHost("host.com"))
		_, path, matches := route.MatchesPath("/test/test")
		assert.True(matches)
		assert.Equal("/test/test", path)
		assert.True(route.MatchesMethod("GET"))
	})
	t.Run("host.com/test2/", func(t *testing.T) {
		route := router.NewRoute("host.com/test2/", nil, func(w http.ResponseWriter, req *http.Request) {}, "POST", "PUT")
		t.Log(route)
		assert.NotNil(route)
		assert.Equal("*.host.com", route.Host())
		assert.Equal("/test2", route.Path())
		assert.Equal([]string{"POST", "PUT"}, route.Methods())
		_, path, matches := route.MatchesPath("/test2")
		assert.True(matches)
		assert.Equal(path, "/")
		_, path2, matches2 := route.MatchesPath("/test2/")
		assert.True(matches2)
		assert.Equal(path2, "/")
		assert.True(route.MatchesMethod("POST"))
		assert.True(route.MatchesMethod("PUT"))
		assert.False(route.MatchesMethod("GET"))
	})
	t.Run("host.com/test/99", func(t *testing.T) {
		route := router.NewRoute("host.com/test", regexp.MustCompile(`^(\d+)`), func(w http.ResponseWriter, req *http.Request) {})
		t.Log(route)
		assert.NotNil(route)
		if params, path, matches := route.MatchesPath("/test/99"); matches {
			assert.Equal(path, "/99")
			assert.Equal([]string{"99"}, params)
		} else {
			assert.Fail("no match")
		}
		if _, _, matches := route.MatchesPath("/test99"); matches {
			assert.Fail("bad match")
		}
	})
}
