package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	client "github.com/mutablelogic/go-client"
	server "github.com/mutablelogic/go-server"
	metric "go.opentelemetry.io/otel/metric"
	trace "go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type global struct {
	// Debug options
	Debug        bool             `name:"debug" help:"Enable debug logging"`
	Verbose      bool             `name:"verbose" help:"Enable verbose logging"`
	PrintVersion kong.VersionFlag `name:"version" help:"Print version and exit"`

	// HTTP server options
	HTTP struct {
		Prefix  string        `name:"prefix" help:"HTTP path prefix" default:"/api"`
		Addr    string        `name:"addr" env:"${ENV_NAME}_ADDR,ADDR" help:"HTTP Listen address" default:"localhost:8084"`
		Timeout time.Duration `name:"timeout" help:"HTTP server read/write timeout" default:"15m"`
	} `embed:"" prefix:"http."`

	// Open Telemetry options
	OTel struct {
		TracesEndpoint  string `env:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT,OTEL_EXPORTER_OTLP_ENDPOINT" help:"Open Telemetry traces endpoint" default:""`
		LogEndpoint     string `env:"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT,OTEL_EXPORTER_OTLP_ENDPOINT" help:"Open Telemetry logs endpoint" default:""`
		MetricsEndpoint string `env:"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT,OTEL_EXPORTER_OTLP_ENDPOINT" help:"Open Telemetry metrics endpoint" default:""`
		Header          string `env:"OTEL_EXPORTER_OTLP_HEADERS" help:"OpenTelemetry collector headers"`
		Name            string `env:"OTEL_SERVICE_NAME" help:"OpenTelemetry service name" default:"${EXECUTABLE_NAME}"`
	} `embed:"" prefix:"otel."`

	// Private fields
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *slog.Logger
	meter       metric.Meter
	tracer      trace.Tracer
	execName    string
	version     string
	description string
	url         *url.URL

	// Defaults store for command defaults management
	defaults
}

// Ensure *globals implements server.Cmd
var _ server.Cmd = (*global)(nil)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (g *global) Name() string {
	return g.execName
}

func (g *global) Description() string {
	return g.description
}

func (g *global) Version() string {
	return g.version
}

func (g *global) Context() context.Context {
	return g.ctx
}

func (g *global) Logger() *slog.Logger {
	return g.logger
}

func (g *global) Tracer() trace.Tracer {
	return g.tracer
}

func (g *global) Meter() metric.Meter {
	return g.meter
}

// ClientEndpoint returns the HTTP endpoint URL and client options derived
// from the global HTTP flags.
func (g *global) ClientEndpoint() (string, []client.ClientOpt, error) {
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

func (g *global) URL() *url.URL {
	if g.url == nil {
		endpoint, _, err := g.ClientEndpoint()
		if err != nil {
			return nil
		}
		if url, err := url.Parse(endpoint); err == nil {
			g.url = url
		}
	}
	return g.url
}

func (g *global) IsTerm() int {
	return TerminalWidth()
}

func (g *global) IsDebug() bool {
	return g.Debug || g.Verbose
}

func (g *global) HTTPAddr() string {
	return g.HTTP.Addr
}

func (g *global) ServerName() string {
	return g.HTTP.Addr
}

func (g *global) HTTPPrefix() string {
	return g.HTTP.Prefix
}

func (g *global) HTTPTimeout() time.Duration {
	return g.HTTP.Timeout
}
