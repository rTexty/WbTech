// Package tracing provides distributed tracing setup using OpenTelemetry.
package tracing

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds tracing configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	Enabled        bool
}

// Tracer wraps OpenTelemetry tracer for application use.
type Tracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	enabled  bool
}

// New initializes OpenTelemetry tracing with the given configuration.
func New(ctx context.Context, cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		log.Println("Tracing is disabled")
		return &Tracer{enabled: false}, nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.JaegerEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("Tracing initialized: service=%s, endpoint=%s", cfg.ServiceName, cfg.JaegerEndpoint)

	return &Tracer{
		tracer:   provider.Tracer(cfg.ServiceName),
		provider: provider,
		enabled:  true,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

// StartSpan starts a new span with the given name.
func (t *Tracer) StartSpan(
	ctx context.Context, name string, opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return t.tracer.Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetError marks the span as errored.
func SetError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
