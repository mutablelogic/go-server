package otel

import (
	"context"
	"fmt"
	"net"
	"net/url"

	// Packages
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func parseEndpoint(endpoint string) (*url.URL, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("missing OTLP endpoint")
	}
	if host, port, err := net.SplitHostPort(endpoint); err == nil && host != "" && port != "" {
		endpoint = "https://" + endpoint
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
