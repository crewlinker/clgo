package clpostgres

import (
	"context"

	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a pgx logger that uses a main zap logger for logging but will prefer
// using a context specific logger if it exists.
type Logger struct {
	dbcfg *pgxpool.Config
}

// NewLogger inits a logger for pgx. Inside is a contextual logger so we can log each postgres query
// with context fields for tracing.
func NewLogger(dbcfg *pgxpool.Config) *Logger {
	return &Logger{
		dbcfg: dbcfg,
	}
}

// Log implements the postgres logger.
func (pl *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, 0, len(data))
	logs := clzap.Log(ctx).WithOptions(zap.AddCallerSkip(1))

	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	fields = append(fields,
		zap.String("db_name", pl.dbcfg.ConnConfig.Database),
		zap.String("db_host", pl.dbcfg.ConnConfig.Host))

	switch level {
	case tracelog.LogLevelTrace:
		logs.Debug(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	case tracelog.LogLevelDebug:
		logs.Debug(msg, fields...)
	case tracelog.LogLevelInfo:
		logs.Info(msg, fields...)
	case tracelog.LogLevelWarn:
		logs.Warn(msg, fields...)
	case tracelog.LogLevelError:
		logs.Error(msg, fields...)
	case tracelog.LogLevelNone:
	default:
		logs.Error(msg, append(fields, zap.Stringer("PGX_LOG_LEVEL", level))...)
	}
}
