package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	// Packages
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	otel "go.opentelemetry.io/otel"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// freeAddr returns a free localhost TCP address by opening a listener on an
// ephemeral port and immediately closing it. There is a small race window
// between close and the server binding, but it is acceptable in tests.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("freeAddr: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

// waitHTTP polls url every 20 ms until it receives any HTTP response (any
// status code is acceptable — a 404 proves the server is serving) or the
// 3-second deadline expires.
func waitHTTP(url string) bool {
	deadline := time.Now().Add(3 * time.Second)
	client := &http.Client{Timeout: 200 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// newTestGlobal builds a *global wired to a cancellable context and a
// discard-logger, ready to pass to RunServer.Run.
func newTestGlobal(t *testing.T, addr string) *global {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	g := &global{
		ctx:      ctx,
		cancel:   cancel,
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
		execName: "test",
		version:  "dev",
	}
	g.HTTP.Addr = addr
	g.HTTP.Prefix = "/api"
	return g
}

///////////////////////////////////////////////////////////////////////////////
// RunServer.Run INTEGRATION TESTS

// Test_RunServer_Run starts a real HTTP server on a pre-allocated port,
// confirms it accepts connections, cancels the context, and asserts a clean
// (nil-error) shutdown.
func Test_RunServer_Run(t *testing.T) {
	addr := freeAddr(t)
	g := newTestGlobal(t, addr)
	s := &RunServer{}

	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(g) }()

	// Any HTTP response proves the server is up (404 is fine)
	if !waitHTTP(fmt.Sprintf("http://%s/api", addr)) {
		g.cancel()
		t.Fatal("server did not accept connections within 3s")
	}

	// Trigger graceful shutdown and check the return value
	g.cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down within 5s after context cancellation")
	}
}

// Test_RunServer_Run_Register verifies that routes added via Register are
// reachable while the server is running.
func Test_RunServer_Run_Register(t *testing.T) {
	addr := freeAddr(t)
	g := newTestGlobal(t, addr)
	s := &RunServer{}
	s.Register(func(router *httprouter.Router) error {
		return router.RegisterFunc("ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, false, nil)
	})

	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(g) }()

	pingURL := fmt.Sprintf("http://%s/api/ping", addr)
	if !waitHTTP(pingURL) {
		g.cancel()
		t.Fatal("/api/ping was not reachable within 3s")
	}

	// Confirm the registered handler returns the expected status code
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(pingURL)
	if err != nil {
		t.Fatalf("GET /api/ping: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("GET /api/ping = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	g.cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down within 5s after context cancellation")
	}
}

///////////////////////////////////////////////////////////////////////////////
// Main — OTel branch INTEGRATION TEST

// Test_Main_OTel exercises the traces-endpoint branch in Main.
// A minimal httptest.Server acts as a fake OTLP collector: it accepts any
// request and returns 200 OK. The test verifies that Main succeeds and the
// OTel provider is initialised and shut down cleanly.
func Test_Main_OTel(t *testing.T) {
	otlpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer otlpSrv.Close()

	type otelCmds struct {
		Run runTestCmd `cmd:"" name:"run" help:"run"`
	}

	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = []string{
		"test",
		"--otel.traces-endpoint", otlpSrv.URL,
		"--otel.name", "test-otel-svc",
		"run",
	}

	if err := Main(otelCmds{}, "OTel integration test", "v0"); err != nil {
		t.Errorf("Main() with --otel.traces-endpoint returned error: %v", err)
	}
}

// Test_RunServer_Run_WithTracer covers the ctx.tracer != nil branch in
// RunServer.Run (which wraps the router with the OTel HTTP middleware).
func Test_RunServer_Run_WithTracer(t *testing.T) {
	addr := freeAddr(t)
	g := newTestGlobal(t, addr)
	g.tracer = otel.Tracer("test-tracer") // noop tracer from default global provider

	s := &RunServer{}
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(g) }()

	if !waitHTTP(fmt.Sprintf("http://%s/api", addr)) {
		g.cancel()
		t.Fatal("server did not accept connections within 3s")
	}

	g.cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run() with OTel tracer returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down within 5s after context cancellation")
	}
}

// Test_RunServer_Run_WithTimeout covers the ctx.HTTP.Timeout > 0 branch in
// RunServer.Run (which adds read/write timeout server options).
func Test_RunServer_Run_WithTimeout(t *testing.T) {
	addr := freeAddr(t)
	g := newTestGlobal(t, addr)
	g.HTTP.Timeout = 30 * time.Second

	s := &RunServer{}
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(g) }()

	if !waitHTTP(fmt.Sprintf("http://%s/api", addr)) {
		g.cancel()
		t.Fatal("server did not accept connections within 3s")
	}

	g.cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run() with timeout returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down within 5s after context cancellation")
	}
}
