package dnsregister_test

import (
	"context"
	"testing"

	// Module import
	dnsregister "github.com/mutablelogic/go-server/pkg/dnsregister"
	task "github.com/mutablelogic/go-server/pkg/task"
	plugin "github.com/mutablelogic/go-server/plugin"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Task_001(t *testing.T) {
	// Create a provider, register dnsregister
	provider, err := task.NewProvider(context.Background(), dnsregister.WithLabel(t.Name()))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(provider)
	}
}

func Test_Task_002(t *testing.T) {
	// Create a provider, register dnsregister
	provider, err := task.NewProvider(context.Background(), dnsregister.WithLabel(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	// Get the task
	task := provider.Get("dnsregister", t.Name())
	if task == nil {
		t.Fatal("Expected task")
	} else {
		t.Log(task)
	}
}

func Test_Task_003(t *testing.T) {
	provider, err := task.NewProvider(context.Background(), dnsregister.WithLabel(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	task, ok := provider.Get("dnsregister", t.Name()).(plugin.DNSRegister)
	if !ok || task == nil {
		t.Fatal("Expected DNSRegister task")
	}
	if ip, err := task.GetExternalAddress(); err != nil {
		t.Error(err)
	} else {
		t.Log("external address=", ip)
	}
}
