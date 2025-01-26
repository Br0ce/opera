package monitor

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// StartTracing bootstraps OpenTelemetry by setting a propagator and a trace provider with
// the given trace provider addr.
// Use the returned shutdown func to gracefully shutdown OpenTelemetry in case of no error.
func StartTracing(ctx context.Context, tpAddr string) (func(context.Context) error, error) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	tp, err := newTraceProvider(tpAddr)
	if err != nil {
		err = errors.Join(err, tp.Shutdown(ctx))
		return nil, fmt.Errorf("set trace provider: %w", err)
	}
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// StartTestTracing returns a noop trace provider for testing and tpAddr is not considered.
// If integration is set to true, tpAddr is forwarded to StartTracing.
func StartTestTracing(ctx context.Context, integration bool, tpAddr string) (func(context.Context) error, error) {
	if integration {
		return StartTracing(ctx, tpAddr)
	}

	exporter := tracetest.NewNoopExporter()
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

// newTraceProvider creates a trace provider with the given addr.
func newTraceProvider(addr string) (*trace.TracerProvider, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(addr),
		// TODO: Enable TLS
		otlptracehttp.WithInsecure(),
	)

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	return tp, nil
}
