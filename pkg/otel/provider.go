package otel

import (
	"context"
	"errors"
	"fmt"
	"sync"

	// Packages
	gootel "go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	lognoop "go.opentelemetry.io/otel/log/noop"
	metric "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	propagation "go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	providerMu     sync.Mutex
	globalProvider *Provider
	propagatorOnce sync.Once
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Provider struct {
	tracer *sdktrace.TracerProvider
	meter  *sdkmetric.MeterProvider
	logger *sdklog.LoggerProvider
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProvider(tracesEndpoint, metricsEndpoint, logsEndpoint, header, name string, attrs ...Attr) (*Provider, error) {
	if tracesEndpoint == "" && metricsEndpoint == "" && logsEndpoint == "" {
		return nil, fmt.Errorf("missing OTLP endpoint")
	}

	res, err := newResource(name, attrs)
	if err != nil {
		return nil, err
	}

	provider := new(Provider)
	if tracesEndpoint != "" {
		provider.tracer, err = newTraceProvider(tracesEndpoint, header, res)
		if err != nil {
			return nil, err
		}
	}
	if metricsEndpoint != "" {
		provider.meter, err = newMeterProvider(metricsEndpoint, header, res)
		if err != nil {
			return nil, err
		}
	}
	if logsEndpoint != "" {
		provider.logger, err = newLoggerProvider(logsEndpoint, header, res)
		if err != nil {
			return nil, err
		}
	}

	providerMu.Lock()
	defer providerMu.Unlock()
	if globalProvider != nil {
		return nil, fmt.Errorf("global OTel provider already set; call ShutdownProvider first")
	}
	if provider.tracer != nil {
		propagatorOnce.Do(func() {
			gootel.SetTextMapPropagator(propagation.TraceContext{})
		})
		gootel.SetTracerProvider(provider.tracer)
	}
	if provider.meter != nil {
		gootel.SetMeterProvider(provider.meter)
	}
	if provider.logger != nil {
		logglobal.SetLoggerProvider(provider.logger)
	}
	globalProvider = provider

	return provider, nil
}

func ShutdownProvider(ctx context.Context) error {
	providerMu.Lock()
	p := globalProvider
	globalProvider = nil
	if p != nil {
		if p.tracer != nil {
			gootel.SetTracerProvider(tracenoop.NewTracerProvider())
		}
		if p.meter != nil {
			gootel.SetMeterProvider(metricnoop.NewMeterProvider())
		}
		if p.logger != nil {
			logglobal.SetLoggerProvider(lognoop.NewLoggerProvider())
		}
	}
	providerMu.Unlock()
	if p == nil {
		return nil
	}
	return p.Shutdown(ctx)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if p == nil || p.tracer == nil {
		return nil
	}
	return p.tracer.Tracer(name, opts...)
}

func (p *Provider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	if p == nil || p.meter == nil {
		return nil
	}
	return p.meter.Meter(name, opts...)
}

func (p *Provider) Logger(name string, opts ...otellog.LoggerOption) otellog.Logger {
	if p == nil || p.logger == nil {
		return nil
	}
	return p.logger.Logger(name, opts...)
}

func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}

	var err error
	if p.logger != nil {
		err = errors.Join(err, p.logger.Shutdown(ctx))
	}
	if p.meter != nil {
		err = errors.Join(err, p.meter.Shutdown(ctx))
	}
	if p.tracer != nil {
		err = errors.Join(err, p.tracer.Shutdown(ctx))
	}
	return err
}
