package cmd

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	// Packages
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	otel "go.opentelemetry.io/otel"
)

// Test_Global_Accessors verifies the simple accessor methods on *global.
func Test_Global_Accessors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	g := &global{
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
	s.Register(func(_ *httprouter.Router) error {
		called = append(called, 1)
		return nil
	})
	s.Register(func(_ *httprouter.Router) error {
		called = append(called, 2)
		return nil
	})

	if len(s.register) != 2 {
		t.Fatalf("expected 2 registered funcs, got %d", len(s.register))
	}

	// Execute them manually to verify ordering
	for _, fn := range s.register {
		if err := fn(nil); err != nil {
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
	got := s.Register(func(_ *httprouter.Router) error { return nil })
	if got != s {
		t.Error("Register() did not return the receiver")
	}
}

// Test_ClientEndpoint_Timeout verifies that a positive Timeout adds an opt.
func Test_ClientEndpoint_Timeout(t *testing.T) {
	g := &global{}
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
	g := &global{}
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
	g := &global{}
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

func Test_RunServer_AdvertisedURL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := &global{
		ctx:    ctx,
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	g.HTTP.Addr = ":0"
	g.HTTP.Prefix = "/api"

	s := &RunServer{}
	var registerURL string
	s.Register(func(_ *httprouter.Router) error {
		if url := g.URL(); url != nil {
			registerURL = url.String()
		}
		return nil
	})

	done := make(chan error, 1)
	go func() {
		done <- s.Run(g)
	}()

	time.Sleep(100 * time.Millisecond)
	if registerURL == "" {
		t.Fatal("expected URL to be available during route registration")
	}
	if !strings.HasPrefix(registerURL, "http://localhost:") {
		t.Fatalf("register URL = %q, want localhost prefix", registerURL)
	}
	if strings.Contains(registerURL, ":0/") {
		t.Fatalf("register URL = %q, want bound port instead of :0", registerURL)
	}

	if url := g.URL(); url == nil {
		t.Fatal("expected URL after listen")
	} else if strings.Contains(url.String(), ":0/") {
		t.Fatalf("final URL = %q, want bound port", url.String())
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func Test_RunServer_AdvertisedURL_TLSName(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := &global{
		ctx:    ctx,
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	g.HTTP.Addr = ":0"
	g.HTTP.Prefix = "/api"

	s := &RunServer{}
	s.TLS.ServerName = "auth.example.com"
	var registerURL string
	s.Register(func(_ *httprouter.Router) error {
		if url := g.URL(); url != nil {
			registerURL = url.String()
		}
		return nil
	})

	done := make(chan error, 1)
	go func() {
		done <- s.Run(g)
	}()

	time.Sleep(100 * time.Millisecond)
	if registerURL != "https://auth.example.com/api" {
		t.Fatalf("register URL = %q, want %q", registerURL, "https://auth.example.com/api")
	}

	if url := g.URL(); url == nil {
		t.Fatal("expected URL after listen")
	} else if got := url.String(); got != "https://auth.example.com/api" {
		t.Fatalf("final URL = %q, want %q", got, "https://auth.example.com/api")
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
