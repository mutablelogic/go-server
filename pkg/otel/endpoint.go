package otel

import (
	"context"
	"fmt"
	"net"
	"net/url"
<<<<<<< HEAD
=======
	"strings"
>>>>>>> f4b5e0d339ba87cf08d93fc3b72c99a1e8e10166

	// Packages
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

<<<<<<< HEAD
=======
// parseEndpoint parses an OTLP endpoint string into a URL.
//
// Callers typically pass values sourced from the command-layer environment
// bindings:
//
//   - traces: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
//   - logs: OTEL_EXPORTER_OTLP_LOGS_ENDPOINT
//   - metrics: OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
//   - fallback for all three: OTEL_EXPORTER_OTLP_ENDPOINT
//
// Accepted endpoint formats are:
//
//   - full HTTP/HTTPS URLs, for example:
//     https://loki.example.com/otlp
//     https://prometheus.example.com/api/v1/otlp
//     https://jaeger.example.com
//   - full gRPC URLs, for example:
//     grpc://otel-collector.internal:4317
//     grpcs://otel-collector.example.com:4317
//   - bare host:port values, for example:
//     localhost:4318
//
// Bare host:port values are treated as HTTPS and normalized to
// https://host:port. Endpoints used with the HTTP exporters may optionally
// include a base path. Signal-specific paths such as /v1/logs, /v1/metrics,
// and /v1/traces are appended later by the respective exporter helpers.
// gRPC/grpcs endpoints must not include a path; the signal-specific RPC method
// is handled by the gRPC exporters themselves.
>>>>>>> f4b5e0d339ba87cf08d93fc3b72c99a1e8e10166
func parseEndpoint(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("missing OTLP endpoint")
	}
<<<<<<< HEAD
	if host, port, err := net.SplitHostPort(endpoint); err == nil && host != "" && port != "" {
		endpoint = "https://" + endpoint
=======
	if !strings.Contains(endpoint, "://") {
		if host, port, err := net.SplitHostPort(endpoint); err == nil && host != "" && port != "" {
			endpoint = "https://" + endpoint
		}
>>>>>>> f4b5e0d339ba87cf08d93fc3b72c99a1e8e10166
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid OTLP endpoint %q: %w", endpoint, err)
	}
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("OTLP endpoint %q missing scheme", endpoint)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("OTLP endpoint %q missing host", endpoint)
	}
	return parsed, nil
}

func newResource(name string, attrs []Attr) (*sdkresource.Resource, error) {
	if name != "" {
		attrs = append(attrs, Attr{Key: "service.name", Value: name})
	}
	return sdkresource.New(
		context.Background(),
		sdkresource.WithAttributes(toAttributes(attrs)...),
		sdkresource.WithHost(),
		sdkresource.WithProcess(),
		sdkresource.WithTelemetrySDK(),
	)
}
