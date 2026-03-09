package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	server "github.com/mutablelogic/go-server"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Global struct {
	// Debug options
	Debug        bool             `name:"debug" help:"Enable debug logging"`
	Verbose      bool             `name:"verbose" help:"Enable verbose logging"`
	PrintVersion kong.VersionFlag `name:"version" help:"Print version and exit"`

	// HTTP server options
	HTTP struct {
		Prefix  string        `name:"prefix" help:"HTTP path prefix" default:"/api"`
		Addr    string        `name:"addr" env:"${ENV_NAME}_ADDR,ADDR" help:"HTTP Listen address" default:"localhost:8084"`
		Timeout time.Duration `name:"timeout" help:"HTTP server read/write timeout" default:"15m"`
		Origin  string        `name:"origin" help:"Cross-origin protection (CSRF) origin. Empty string for same-origin only, '*' to allow all cross-origin requests, or a specific origin in the form 'scheme://host[:port]'." default:""`
	} `embed:"" prefix:"http."`

	// Open Telemetry options
	OTel struct {
		Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" help:"Open Telemetry endpoint" default:""`
		Header   string `env:"OTEL_EXPORTER_OTLP_HEADERS" help:"OpenTelemetry collector headers"`
		Name     string `env:"OTEL_SERVICE_NAME" help:"OpenTelemetry service name" default:"${EXECUTABLE_NAME}"`
	} `embed:"" prefix:"otel."`

	// Private fields
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger
	tracer      trace.Tracer
	execName    string
	version     string
	description string
}

// Ensure *Global implements server.Cmd
var _ server.Cmd = (*Global)(nil)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (g *Global) Name() string {
	return g.execName
}

func (g *Global) Description() string {
	return g.description
}

func (g *Global) Version() string {
	return g.version
}

func (g *Global) Context() context.Context {
	return g.ctx
}

func (g *Global) Logger() *slog.Logger {
	return g.logger
}

func (g *Global) Tracer() trace.Tracer {
	return g.tracer
}

// ClientEndpoint returns the HTTP endpoint URL and client options derived
// from the global HTTP flags.
func (g *Global) ClientEndpoint() (string, []client.ClientOpt, error) {
	scheme := "http"
	host, port, err := net.SplitHostPort(g.HTTP.Addr)
	if err != nil {
		return "", nil, err
	}
	if host == "" {
		host = "localhost"
	}
	portn, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return "", nil, err
	}
	if portn == 443 {
		scheme = "https"
	}
	opts := []client.ClientOpt{}
	if g.Debug || g.Verbose {
		opts = append(opts, client.OptTrace(os.Stderr, g.Verbose))
	}
	if g.tracer != nil {
		opts = append(opts, client.OptTracer(g.tracer))
	}
	if g.HTTP.Timeout > 0 {
		opts = append(opts, client.OptTimeout(g.HTTP.Timeout))
	}
	return fmt.Sprintf("%s://%s%s", scheme, net.JoinHostPort(host, strconv.FormatUint(portn, 10)), g.HTTP.Prefix), opts, nil
}
