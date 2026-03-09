package cmd

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	otel "go.opentelemetry.io/otel"
)

// Test_Global_Accessors verifies the simple accessor methods on *Global.
func Test_Global_Accessors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	g := &Global{
		execName: "testbin",
		version:  "v9.9.9",
		ctx:      ctx,
		logger:   logger,
	}

	if got := g.Name(); got != "testbin" {
		t.Errorf("Name() = %q, want %q", got, "testbin")
	}
	if got := g.Version(); got != "v9.9.9" {
		t.Errorf("Version() = %q, want %q", got, "v9.9.9")
	}
	if got := g.Context(); got != ctx {
		t.Error("Context() did not return the expected context")
	}
	if got := g.Logger(); got != logger {
		t.Error("Logger() did not return the expected logger")
	}
	if got := g.Tracer(); got != nil {
		t.Errorf("Tracer() = %v, want nil", got)
	}
}

// Test_RunServer_Register verifies that Register accumulates funcs and they
// are called in order.
func Test_RunServer_Register(t *testing.T) {
	var called []int

	s := &RunServer{}
	s.Register(func(_ *httprouter.Router, _ server.Cmd) error {
		called = append(called, 1)
		return nil
	})
	s.Register(func(_ *httprouter.Router, _ server.Cmd) error {
		called = append(called, 2)
		return nil
	})

	if len(s.register) != 2 {
		t.Fatalf("expected 2 registered funcs, got %d", len(s.register))
	}

	// Execute them manually to verify ordering
	for _, fn := range s.register {
		if err := fn(nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(called) != 2 || called[0] != 1 || called[1] != 2 {
		t.Errorf("funcs called out of order or missing: %v", called)
	}
}

// Test_RunServer_Register_Chaining verifies the fluent chaining return.
func Test_RunServer_Register_Chaining(t *testing.T) {
	s := &RunServer{}
	got := s.Register(func(_ *httprouter.Router, _ server.Cmd) error { return nil })
	if got != s {
		t.Error("Register() did not return the receiver")
	}
}

// Test_ClientEndpoint_Timeout verifies that a positive Timeout adds an opt.
func Test_ClientEndpoint_Timeout(t *testing.T) {
	g := &Global{}
	g.HTTP.Addr = "localhost:8084"
	g.HTTP.Prefix = "/api"
	g.HTTP.Timeout = 30 * time.Second

	_, opts, err := g.ClientEndpoint()
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) == 0 {
		t.Error("expected opts when Timeout > 0, got none")
	}
}

// Test_ClientEndpoint_Tracer verifies that a non-nil tracer adds an opt.
func Test_ClientEndpoint_Tracer(t *testing.T) {
	g := &Global{}
	g.HTTP.Addr = "localhost:8084"
	g.HTTP.Prefix = "/api"
	g.tracer = otel.Tracer("test") // noop tracer via default global provider

	_, opts, err := g.ClientEndpoint()
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) == 0 {
		t.Error("expected opts when tracer != nil, got none")
	}
}

// Test_ClientEndpoint_Verbose verifies that Verbose=true adds an opt.
func Test_ClientEndpoint_Verbose(t *testing.T) {
	g := &Global{}
	g.HTTP.Addr = "localhost:8084"
	g.HTTP.Prefix = "/api"
	g.Verbose = true

	_, opts, err := g.ClientEndpoint()
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) == 0 {
		t.Error("expected opts when Verbose=true, got none")
	}
}
