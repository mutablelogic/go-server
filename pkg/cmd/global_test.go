package cmd

import (
	"os"
	"testing"
)

type runTestCmd struct{}

func (r *runTestCmd) Run(*Global) error { return nil }

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
			g := &Global{}
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

// Test_Main_OTelError verifies that an invalid OTel endpoint causes Main to
// return an error (exercises the otel.NewProvider failure path).
func Test_Main_OTelError(t *testing.T) {
	type errCmds struct {
		Run runTestCmd `cmd:"" name:"run" help:"run"`
	}

	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = []string{"test", "--otel.endpoint", "not-a-valid-url", "run"}

	if err := Main(errCmds{}, "OTel error test", "v0"); err == nil {
		t.Error("Main() expected an error for invalid OTel endpoint, got nil")
	}
}

func Test_ClientEndpoint_Opts(t *testing.T) {
	g := &Global{}
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
