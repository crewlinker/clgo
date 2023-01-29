package clpostgres

import (
	"context"

	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a pgx logger that uses a main zap logger for logging but will prefer
// using a context specific logger if it exists
type Logger struct {
	logs *clzap.ContextLogger
}

// NewLogger inits a logger for pgx. Inside is a contextual logger so we can log each postgres query
// with context fields for tracing.
func NewLogger(logs *zap.Logger) *Logger {
	return &Logger{logs: clzap.NewTraceContextLogger(logs.WithOptions(zap.AddCallerSkip(1)))}
}

func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, len(data))
	i := 0
	for k, v := range data {
		fields[i] = zap.Any(k, v)
		i++
	}

	switch level {
	case tracelog.LogLevelTrace:
		pl.logs.Debug(ctx, msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case tracelog.LogLevelDebug:
		pl.logs.Debug(ctx, msg, fields...)
	case tracelog.LogLevelInfo:
		pl.logs.Info(ctx, msg, fields...)
	case tracelog.LogLevelWarn:
		pl.logs.Warn(ctx, msg, fields...)
	case tracelog.LogLevelError:
		pl.logs.Error(ctx, msg, fields...)
	default:
		pl.logs.Error(ctx, msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}
