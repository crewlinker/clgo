package clzap

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ctxKey holds the context key under which the logger will be stored.
type ctxKey string

// Log retrieves a zap logger from the context. Returns a no-op logger if none is defined. If the context also
// has tracing and or span information this will be logged by the logger automatically.
func Log(ctx context.Context) *zap.Logger {
	logs, ok := ctx.Value(ctxKey("clzap.logger")).(*zap.Logger)
	if !ok {
		logs = zap.NewNop()
	}

	// if span information is in the context, add it as a field to the logger
	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().HasSpanID() {
		logs = logs.With(zap.String("span_id", span.SpanContext().SpanID().String()))
	}

	// log the trace id in the xray format, and add it as a field to the logger
	if span != nil && span.SpanContext().HasTraceID() {
		tid := span.SpanContext().TraceID().String()
		logs = logs.With(zap.String("trace_id", fmt.Sprintf("1-%s-%s", tid[:8], tid[8:])))
	}

	return logs
}

// WithLogger returns a context with the provided logger embedded.
func WithLogger(ctx context.Context, logs *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey("clzap.logger"), logs)
}
