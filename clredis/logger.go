package clredis

import (
	"context"
	"log"

	"go.uber.org/zap"
)

// Logger logger implements: https://github.com/go-redis/redis/blob/86258a11a93291b817f420fff47e79bd22fcb4ca/redis.go
type Logger struct{ logs *log.Logger }

// NewLogger inits a Redis logger from a zap logger.
func NewLogger(logs *zap.Logger) *Logger {
	return &Logger{zap.NewStdLog(logs)}
}

// Printf implements the redis logger.
func (l Logger) Printf(_ context.Context, format string, v ...interface{}) {
	l.logs.Printf(format, v...)
}
