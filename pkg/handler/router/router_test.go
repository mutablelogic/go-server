package router_test

import (
	"context"
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
	assert.NoError(router.AddHandler(context.Background(), "/hello", nil))
}

func Test_router_002(t *testing.T) {
	assert := assert.New(t)
	task, err := router.Config{}.New(context.Background())
	assert.NoError(err)
	assert.NotNil(task)
	router := task.(server.Router)

	path := regexp.MustCompile("^/hello$")
	assert.NoError(router.AddHandlerRe(context.Background(), "mutablelogic.com", path, nil))
	assert.NoError(router.AddHandler(context.Background(), "mutablelogic.com/", nil))

	results := router.Match("mutablelogic.com", "GET", "/hello")
	assert.NotNil(results)
	t.Log(results)

	results = router.Match("mutablelogic.com", "GET", "/")
	assert.NotNil(results)
	t.Log(results)
}
