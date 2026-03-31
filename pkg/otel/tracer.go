package otel

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	// Packages
	gootel "go.opentelemetry.io/otel"
	attribute "go.opentelemetry.io/otel/attribute"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	propagation "go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type Attr struct {
	Key   string
	Value string
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewTracer(endpoint, header, name string, attrs ...Attr) (func(context.Context) error, error) {
	res, err := newResource(name, attrs)
	if err != nil {
		return nil, err
	}

	provider, err := newTraceProvider(endpoint, header, res)
	if err != nil {
		return nil, err
	}

	propagatorOnce.Do(func() {
		gootel.SetTextMapPropagator(propagation.TraceContext{})
	})
	gootel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newTraceProvider(endpoint, header string, res *sdkresource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := newTraceExporter(endpoint, header)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}

// Provide a tracing exporter based on the provided endpoint and header configuration.
// The endpoint can be in the form of host:port for HTTPS endpoints, or a
// URL with http, https, grpc or grpcs scheme, host, port and optional path.
func newTraceExporter(endpoint, header string) (sdktrace.SpanExporter, error) {
	parsed, err := parseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	headers := toHeaders(header)
	switch parsed.Scheme {
	case "http", "https":
		return toHTTP(parsed, headers)
	case "grpc", "grpcs":
		return toGRPC(parsed, headers)
	default:
		return nil, fmt.Errorf("unsupported OTLP scheme %q", parsed.Scheme)
	}
}

func toHTTP(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
	clientOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint.Host),
	}

	if endpoint.Scheme == "http" {
		clientOpts = append(clientOpts, otlptracehttp.WithInsecure())
	}

	pathComponent := strings.Trim(endpoint.Path, "/")
	switch {
	case pathComponent == "":
		pathComponent = "v1/traces"
	case strings.HasSuffix(pathComponent, "v1/traces"):
		// leave as-is
	default:
		pathComponent = pathComponent + "/v1/traces"
	}
	clientOpts = append(clientOpts, otlptracehttp.WithURLPath("/"+pathComponent))

	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlptracehttp.WithHeaders(headers))
	}

	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func toGRPC(endpoint *url.URL, headers map[string]string) (sdktrace.SpanExporter, error) {
	if endpoint.Path != "" && endpoint.Path != "/" {
		return nil, fmt.Errorf("gRPC OTLP endpoint should not include a path: %q", endpoint.Path)
	}

	clientOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint.Host),
	}

	if endpoint.Scheme == "grpc" {
		clientOpts = append(clientOpts, otlptracegrpc.WithInsecure())
	}

	if len(headers) > 0 {
		clientOpts = append(clientOpts, otlptracegrpc.WithHeaders(headers))
	}

	exporter, err := otlptrace.New(context.Background(), otlptracegrpc.NewClient(clientOpts...))
	if err != nil {
		return nil, err
	}
	return exporter, nil
}

func toHeaders(header string) map[string]string {
	headers := make(map[string]string)
	if header == "" {
		return headers
	}
	for _, pair := range strings.Split(header, ",") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if key := strings.TrimSpace(kv[0]); key != "" {
			headers[key] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

func toAttributes(attrs []Attr) []attribute.KeyValue {
	var result []attribute.KeyValue
	for _, attr := range attrs {
		result = append(result, attribute.String(attr.Key, attr.Value))
	}
	return result
}
