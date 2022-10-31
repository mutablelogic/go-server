package httpserver_test

import (
	"context"
	"testing"
	"time"

	// Module import
	httpserver "github.com/mutablelogic/terraform-provider-nginx/pkg/httpserver"
	provider "github.com/mutablelogic/terraform-provider-nginx/pkg/provider"
)

/////////////////////////////////////////////////////////////////////
// TESTS

func Test_Server_001(t *testing.T) {
	// Create a provider, register http server and router
	provider := provider.New()
	server, err := provider.New(context.Background(), httpserver.Config{})
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(server)
	}
}

func Test_Server_002(t *testing.T) {
	// Create a provider, register http server
	provider := provider.New()
	server, err := provider.New(context.Background(), httpserver.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Run(ctx); err != nil {
		t.Error(err)
	}
}
