package clpostgres

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a pgx logger that uses a main zap logger for logging but will prefer
// using a context specific logger if it exists
type Logger struct {
	logs *zap.Logger
}

// NewLogger inits a logger for the
func NewLogger(logs *zap.Logger) *Logger {
	return &Logger{logs: logs.WithOptions(zap.AddCallerSkip(1))}
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
		pl.logs.Debug(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case tracelog.LogLevelDebug:
		pl.logs.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		pl.logs.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		pl.logs.Warn(msg, fields...)
	case tracelog.LogLevelError:
		pl.logs.Error(msg, fields...)
	default:
		pl.logs.Error(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}
