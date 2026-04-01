package otel

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	// Packages
	otlploggrpc "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otlploghttp "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	global "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewLogger(endpoint, header, name string, attrs ...Attr) (func(context.Context) error, error) {
	res, err := newResource(name, attrs)
	if err != nil {
		return nil, err
	}

	provider, err := newLoggerProvider(endpoint, header, res)
	if err != nil {
		return nil, err
	}

	global.SetLoggerProvider(provider)
	return provider.Shutdown, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newLoggerProvider(endpoint, header string, res *sdkresource.Resource) (*sdklog.LoggerProvider, error) {
	exporter, err := newLogExporter(endpoint, header)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	), nil
}

func newLogExporter(endpoint, header string) (sdklog.Exporter, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	headers := toHeaders(header)
	switch parsed.Scheme {
	case "http", "https":
		return toHTTPLog(parsed, headers)
	case "grpc", "grpcs":
		return toGRPCLog(parsed, headers)
	default:
		return nil, fmt.Errorf("unsupported OTLP scheme %q", parsed.Scheme)
	}
}

func toHTTPLog(endpoint *url.URL, headers map[string]string) (sdklog.Exporter, error) {
	clientOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint.Host),
	}
	if endpoint.Scheme == "http" {
		clientOpts = append(clientOpts, otlploghttp.WithInsecure())
	}
	pathComponent := strings.Trim(endpoint.Path, "/")
	switch {
	case pathComponent == "":
		pathComponent = "v1/logs"
	case strings.HasSuffix(pathComponent, "v1/logs"):
		// leave as-is
	default:
		pathComponent = pathComponent + "/v1/logs"
	}
	clientOpts = append(clientOpts, otlploghttp.WithURLPath("/"+pathComponent))
	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlploghttp.WithHeaders(headers))
	}
	return otlploghttp.New(context.Background(), clientOpts...)
}

func toGRPCLog(endpoint *url.URL, headers map[string]string) (sdklog.Exporter, error) {
	if endpoint.Path != "" && endpoint.Path != "/" {
		return nil, fmt.Errorf("gRPC OTLP endpoint should not include a path: %q", endpoint.Path)
	}
	clientOpts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint.Host),
	}
	if endpoint.Scheme == "grpc" {
		clientOpts = append(clientOpts, otlploggrpc.WithInsecure())
	}
	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlploggrpc.WithHeaders(headers))
	}
	return otlploggrpc.New(context.Background(), clientOpts...)
}
