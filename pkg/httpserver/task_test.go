package httpserver_test

import (
	"context"
	"testing"

	// Module import
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Task_001(t *testing.T) {
	// Create a provider, register http server and router
	provider, err := task.NewProvider(context.Background(), httpserver.Plugin{}.WithLabel("label"))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}
}
