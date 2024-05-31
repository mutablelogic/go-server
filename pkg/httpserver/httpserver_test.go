package httpserver_test

import (
<<<<<<< HEAD
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/stretchr/testify/assert"
)

func Test_httpserver_001(t *testing.T) {
	assert := assert.New(t)
	config := httpserver.Config{}
	assert.NotEmpty(config.Name())
	assert.NotEmpty(config.Description())
}

func Test_httpserver_002(t *testing.T) {
	assert := assert.New(t)
	config := httpserver.Config{}
	server, err := config.New()
	assert.NoError(err)
	assert.NotNil(server)
	t.Log(server)
=======
	"context"
	"path/filepath"
	"testing"
	"time"

	// Module import
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Task_001(t *testing.T) {
	// Create a provider, register http server and router
	provider, err := task.NewProvider(context.Background(), httpserver.WithLabel(t.Name()))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}
}

func Test_Task_002(t *testing.T) {
	plugin := httpserver.Plugin{
		Plugin: task.WithLabel("httpserver", t.Name()),
	}.WithListen(filepath.Join(t.TempDir(), "httpserver.sock"))

	// Create a provider, register http server and router
	provider, err := task.NewProvider(context.Background(), plugin)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}

	// Start FCGI server for one second, then quit
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := provider.Run(ctx); err != nil {
		t.Fatal(err)
	}
>>>>>>> a486469478ac5f8553b60a97d8eb2a7a976d11bd
}
