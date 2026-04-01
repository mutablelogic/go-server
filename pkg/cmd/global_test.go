package cmd

import (
	"os"
	"testing"

	server "github.com/mutablelogic/go-server"
)

type runTestCmd struct{}

func (r *runTestCmd) Run(server.Cmd) error { return nil }

func Test_Main_001(t *testing.T) {
	type testCmds struct {
		Run runTestCmd `cmd:"" name:"run" help:"Run test command."`
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"run", []string{"test", "run"}, false},
		{"run --debug", []string{"test", "--debug", "run"}, false},
		{"run --verbose", []string{"test", "--verbose", "run"}, false},
		{"run --http.addr", []string{"test", "--http.addr", ":9090", "run"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := os.Args
			t.Cleanup(func() { os.Args = orig })
			os.Args = tt.args

			if err := Main(testCmds{}, "Test CLI", "v1.0.0"); (err != nil) != tt.wantErr {
				t.Fatalf("Main() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ClientEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		addr         string
		prefix       string
		wantEndpoint string
		wantErr      bool
	}{
		{
			name:         "ipv4 basic",
			addr:         "localhost:8084",
			prefix:       "/api",
			wantEndpoint: "http://localhost:8084/api",
		},
		{
			name:         "empty host defaults to localhost",
			addr:         ":8084",
			prefix:       "/api",
			wantEndpoint: "http://localhost:8084/api",
		},
		{
			name:         "port 443 uses https",
			addr:         "example.com:443",
			prefix:       "/api",
			wantEndpoint: "https://example.com:443/api",
		},
		{
			name:         "ipv6 address",
			addr:         "[::1]:8084",
			prefix:       "/api",
			wantEndpoint: "http://[::1]:8084/api",
		},
		{
			name:         "ipv6 with port 443",
			addr:         "[::1]:443",
			prefix:       "/api",
			wantEndpoint: "https://[::1]:443/api",
		},
		{
			name:         "custom prefix",
			addr:         "localhost:9000",
			prefix:       "/v2",
			wantEndpoint: "http://localhost:9000/v2",
		},
		{
			name:    "invalid addr",
			addr:    "not-valid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &global{}
			g.HTTP.Addr = tt.addr
			g.HTTP.Prefix = tt.prefix

			endpoint, opts, err := g.ClientEndpoint()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ClientEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if endpoint != tt.wantEndpoint {
				t.Errorf("ClientEndpoint() endpoint = %q, want %q", endpoint, tt.wantEndpoint)
			}
			if len(opts) != 0 {
				t.Errorf("ClientEndpoint() expected no opts, got %d", len(opts))
			}
		})
	}
}

// Test_Main_OTelError verifies that an invalid OTel traces endpoint causes Main to
// return an error (exercises the otel.NewProvider failure path).
func Test_Main_OTelError(t *testing.T) {
	type errCmds struct {
		Run runTestCmd `cmd:"" name:"run" help:"run"`
	}

	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = []string{"test", "--otel.traces-endpoint", "not-a-valid-url", "run"}

	if err := Main(errCmds{}, "OTel error test", "v0"); err == nil {
		t.Error("Main() expected an error for invalid OTel traces endpoint, got nil")
	}
}

func Test_ClientEndpoint_Opts(t *testing.T) {
	g := &global{}
	g.HTTP.Addr = "localhost:8084"
	g.HTTP.Prefix = "/api"
	g.Debug = true

	_, opts, err := g.ClientEndpoint()
	if err != nil {
		t.Fatal(err)
	}
	// Debug=true should add OptTrace
	if len(opts) == 0 {
		t.Error("expected opts when Debug=true, got none")
	}
}

func Test_Defaults_SetGetDelete(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/defaults.json"

	var d defaults
	if err := d.init(path); err != nil {
		t.Fatal(err)
	}

	// Set a value and retrieve it
	if err := d.Set("foo", "bar"); err != nil {
		t.Fatal(err)
	}
	if got := d.GetString("foo"); got != "bar" {
		t.Errorf("got %q, want %q", got, "bar")
	}

	// Missing key returns empty string
	if got := d.GetString("missing"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}

	// Keys returns the correct set
	keys := d.Keys()
	if len(keys) != 1 || keys[0] != "foo" {
		t.Errorf("unexpected keys: %v", keys)
	}

	// Delete via nil
	if err := d.Set("foo", nil); err != nil {
		t.Fatal(err)
	}
	if got := d.GetString("foo"); got != "" {
		t.Errorf("expected empty after delete, got %q", got)
	}
	if len(d.Keys()) != 0 {
		t.Errorf("expected no keys after delete, got %v", d.Keys())
	}
}

func Test_Defaults_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/defaults.json"

	// Write via first instance
	var d1 defaults
	if err := d1.init(path); err != nil {
		t.Fatal(err)
	}
	if err := d1.Set("key", "value"); err != nil {
		t.Fatal(err)
	}

	// Reload via second instance and check value is present
	var d2 defaults
	if err := d2.init(path); err != nil {
		t.Fatal(err)
	}
	if got := d2.GetString("key"); got != "value" {
		t.Errorf("got %q after reload, want %q", got, "value")
	}
}

func Test_Defaults_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/defaults.json"

	// Write garbage JSON
	if err := os.WriteFile(path, []byte("not json {{"), 0600); err != nil {
		t.Fatal(err)
	}

	// init should succeed, discarding the corrupt file
	var d defaults
	if err := d.init(path); err != nil {
		t.Fatalf("init with corrupt file returned error: %v", err)
	}
	if len(d.Keys()) != 0 {
		t.Errorf("expected empty store after corrupt file, got %v", d.Keys())
	}

	// Corrupt file should have been removed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected corrupt file to be removed")
	}
}

func Test_Defaults_UninitializedSet(t *testing.T) {
	// Set on a zero-value defaults (never init'd) must not panic
	var d defaults
	if err := d.Set("k", "v"); err == nil {
		// save will fail because d.path is empty — that's fine, no panic
		_ = err
	}
}
