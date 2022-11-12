package dnsregister_test

import (
	"context"
	"testing"

	// Module import
	dnsregister "github.com/mutablelogic/go-server/pkg/dnsregister"
	task "github.com/mutablelogic/go-server/pkg/task"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Task_001(t *testing.T) {
	// Create a provider, register dnsregister
	provider, err := task.NewProvider(context.Background(), dnsregister.Plugin{}.WithLabel("label"))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}
}
