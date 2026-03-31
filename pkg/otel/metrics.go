package otel

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	// Packages
	gootel "go.opentelemetry.io/otel"
	otlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otlpmetrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewMeter(endpoint, header, name string, attrs ...Attr) (func(context.Context) error, error) {
	res, err := newResource(name, attrs)
	if err != nil {
		return nil, err
	}

	provider, err := newMeterProvider(endpoint, header, res)
	if err != nil {
		return nil, err
	}

	gootel.SetMeterProvider(provider)
	return provider.Shutdown, nil
}

func newMeterProvider(endpoint, header string, res *sdkresource.Resource) (*sdkmetric.MeterProvider, error) {
	exporter, err := newMetricExporter(endpoint, header)
	if err != nil {
		return nil, err
	}

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(res),
	), nil
}

func newMetricExporter(endpoint, header string) (sdkmetric.Exporter, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	headers := toHeaders(header)
	switch parsed.Scheme {
	case "http", "https":
		return toHTTPMetric(parsed, headers)
	case "grpc", "grpcs":
		return toGRPCMetric(parsed, headers)
	default:
		return nil, fmt.Errorf("unsupported OTLP scheme %q", parsed.Scheme)
	}
}

func toHTTPMetric(endpoint *url.URL, headers map[string]string) (sdkmetric.Exporter, error) {
	clientOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint.Host),
	}
	if endpoint.Scheme == "http" {
		clientOpts = append(clientOpts, otlpmetrichttp.WithInsecure())
	}
	pathComponent := strings.Trim(endpoint.Path, "/")
	switch {
	case pathComponent == "":
		pathComponent = "v1/metrics"
	case strings.HasSuffix(pathComponent, "v1/metrics"):
		// leave as-is
	default:
		pathComponent = pathComponent + "/v1/metrics"
	}
	clientOpts = append(clientOpts, otlpmetrichttp.WithURLPath("/"+pathComponent))
	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlpmetrichttp.WithHeaders(headers))
	}
	return otlpmetrichttp.New(context.Background(), clientOpts...)
}

func toGRPCMetric(endpoint *url.URL, headers map[string]string) (sdkmetric.Exporter, error) {
	if endpoint.Path != "" && endpoint.Path != "/" {
		return nil, fmt.Errorf("gRPC OTLP endpoint should not include a path: %q", endpoint.Path)
	}
	clientOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint.Host),
	}
	if endpoint.Scheme == "grpc" {
		clientOpts = append(clientOpts, otlpmetricgrpc.WithInsecure())
	}
	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlpmetricgrpc.WithHeaders(headers))
	}
	return otlpmetricgrpc.New(context.Background(), clientOpts...)
}
