package router_test

import (
	"context"
	"regexp"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_route_001(t *testing.T) {
	assert := assert.New(t)

	ctx := provider.WithLabel(context.Background(), "test")
	route := router.NewRoute(ctx, "mutablelogic.com", "/api/v1", "GET")
	assert.NotNil(route)
	assert.Equal("test", route.Label())
	assert.Equal("mutablelogic.com", route.Host())
	assert.Equal("/api/v1", route.Prefix())
	assert.Equal("", route.Path())
	assert.Equal("GET", route.Methods()[0])
}

func Test_route_002(t *testing.T) {
	assert := assert.New(t)

	route := router.NewRouteWithPath(context.Background(), "mutablelogic.com", "/api/v1", "/test", "GET", "POST")
	assert.NotNil(route)
	assert.Equal("mutablelogic.com", route.Host())
	assert.Equal("/api/v1", route.Prefix())
	assert.Equal("/test", route.Path())
	assert.ElementsMatch([]string{"GET", "POST"}, route.Methods())
}

func Test_route_003(t *testing.T) {
	assert := assert.New(t)

	route := router.NewRouteWithRegexp(context.Background(), "mutablelogic.com", "/api/v1", regexp.MustCompile("/test"))
	assert.NotNil(route)
	assert.Equal("mutablelogic.com", route.Host())
	assert.Equal("/api/v1", route.Prefix())
	assert.Equal("", route.Path())
	assert.ElementsMatch([]string{}, route.Methods())
}

func Test_route_004(t *testing.T) {
	assert := assert.New(t)

	ctx := provider.WithLabel(context.Background(), t.Name())
	route := router.NewRoute(ctx, "mutablelogic.com", "/api/v1", "GET").SetScope("test").SetScope("test", "test2")
	assert.NotNil(route)
	assert.ElementsMatch([]string{"test", "test2"}, route.Scopes())
}
