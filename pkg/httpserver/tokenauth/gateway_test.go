package tokenauth_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver/router"
	"github.com/mutablelogic/go-server/pkg/httpserver/tokenauth"
	"github.com/mutablelogic/go-server/pkg/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

func Test_gateway_001(t *testing.T) {
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
	t.Log(provider)
}
