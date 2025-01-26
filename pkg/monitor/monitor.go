package monitor

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "opera"
	defaultTraceID = "xxxx"
)

func TraceID(span trace.Span) string {
	return span.SpanContext().TraceID().String()
}

func TraceIDFromCtx(ctx context.Context) string {
	if spanContext := trace.SpanFromContext(ctx).SpanContext(); spanContext.IsValid() {
		return spanContext.TraceID().String()
	}
	return defaultTraceID
}
