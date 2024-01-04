package clconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/zap"
)

// Logger logs RPC calls as an interceptor.
type Logger struct {
	cfg  Config
	logs *zap.Logger

	connect.Interceptor
}

// NewLogger inits the logger.
func NewLogger(cfg Config, logs *zap.Logger) *Logger {
	lgr := &Logger{cfg: cfg, logs: logs}
	lgr.Interceptor = connect.UnaryInterceptorFunc(lgr.intercept)

	return lgr
}

func (l Logger) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		clzap.Log(ctx, l.logs).Info("handling request",
			zap.String("peer_add", req.Peer().Addr),
			zap.Any("peer_query", req.Peer().Query),
			zap.String("peer_protocol", req.Peer().Protocol),
			zap.Any("header", req.Header()),
			zap.String("http_method", req.HTTPMethod()),
			zap.String("proc", req.Spec().Procedure))

		resp, err := next(ctx, req)
		if err == nil {
			return resp, nil // nothing to do
		}

		clzap.Log(ctx, l.logs).Error("server error", zap.Error(err), zap.Stack("stack"))

		var cerr *connect.Error
		if !errors.As(err, &cerr) {
			cerr = connect.NewError(connect.CodeUnknown, err)
		}

		// dd debug detail information
		if !l.cfg.DisableStackTraceErrorDetails {
			addStackTraceDebugDetails(ctx, cerr)
		}

		return resp, cerr
	})
}
