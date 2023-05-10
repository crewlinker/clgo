package clzap

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ContextHook allows adding fields based on the context.
type ContextHook func(ctx context.Context, f []zap.Field) []zap.Field

// ContextLogger wraps a zap loger but only allows logging with a context. This allows users to force the
// logging of contextual fields such as trace-ids that would otherwise cause logs to not be observable per
// trace. It is not provided as a dependency to fx because it is expected that this logger is initialized
// inside the components on a case-by-case basis. We want the signature of components constructs NOT to
// be dependant on a contextual logger, to imporove portability.
//
// Deprecated: the less surprising way to do contextual logging is to embed the logger in the context. This approach
// can be found in the ctx.go. It is also simpler and doesn't require a home-made type. The ContextLogger was mainly
// useful to allow named loggers per component.
type ContextLogger struct {
	logs *zap.Logger
	hook ContextHook
}

// TraceHook creates a context logger hook that appends trace information.
func TraceHook() func(ctx context.Context, f []zap.Field) []zap.Field {
	return func(ctx context.Context, field []zap.Field) []zap.Field {
		span := trace.SpanFromContext(ctx)
		if span != nil && span.SpanContext().HasSpanID() {
			field = append(field, zap.String("span_id", span.SpanContext().SpanID().String()))
		}

		// log the trace id in the xray format
		if span != nil && span.SpanContext().HasTraceID() {
			tid := span.SpanContext().TraceID().String()
			field = append(field, zap.String("trace_id", fmt.Sprintf("1-%s-%s", tid[:8], tid[8:])))
		}

		return field
	}
}

// NewTraceContextLogger inits a contextual logger that adds trace information to each log line.
func NewTraceContextLogger(logs *zap.Logger) *ContextLogger {
	return NewContextLogger(logs, TraceHook())
}

// NewContextLogger inits our contextual with the underlying zapcore logger.
func NewContextLogger(logs *zap.Logger, hook ...ContextHook) *ContextLogger {
	l := &ContextLogger{logs: logs, hook: func(ctx context.Context, f []zap.Field) []zap.Field { return f }}
	if len(hook) > 0 {
		l.hook = hook[0]
	}

	return l
}

func (log *ContextLogger) clone(l *zap.Logger) *ContextLogger {
	return &ContextLogger{logs: l, hook: log.hook}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (log *ContextLogger) Named(s string) *ContextLogger {
	return log.clone(log.logs.Named(s))
}

// WithOptions clones the current Logger, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func (log *ContextLogger) WithOptions(opts ...zap.Option) *ContextLogger {
	return log.clone(log.logs.WithOptions(opts...))
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (log *ContextLogger) With(fields ...zap.Field) *ContextLogger {
	return log.clone(log.logs.With(fields...))
}

// Level reports the minimum enabled level for this logger.
//
// For NopLoggers, this is [zapcore.InvalidLevel].
func (log *ContextLogger) Level() zapcore.Level {
	return log.logs.Level()
}

// Check returns a CheckedEntry if logging a message at the specified level
// is enabled. It's a completely optional optimization; in high-performance
// applications, Check can help avoid allocating a slice to hold fields.
func (log *ContextLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.logs.Check(lvl, msg)
}

// Log logs a message at the specified level. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
func (log *ContextLogger) Log(ctx context.Context, lvl zapcore.Level, msg string, fields ...zap.Field) {
	log.logs.Log(lvl, msg, log.hook(ctx, fields)...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *ContextLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Debug(msg, log.hook(ctx, fields)...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *ContextLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Info(msg, log.hook(ctx, fields)...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *ContextLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Warn(msg, log.hook(ctx, fields)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *ContextLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Error(msg, log.hook(ctx, fields)...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (log *ContextLogger) DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.DPanic(msg, log.hook(ctx, fields)...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *ContextLogger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Panic(msg, log.hook(ctx, fields)...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *ContextLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	log.logs.Fatal(msg, log.hook(ctx, fields)...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (log *ContextLogger) Sync() error {
	if err := log.logs.Sync(); err != nil {
		return fmt.Errorf("failed to sync underlying: %w", err)
	}

	return nil
}
