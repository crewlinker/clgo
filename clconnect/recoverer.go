package clconnect

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clzap"
	"github.com/go-stack/stack"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// ErrServerPanic is returned when the server panicked.
var ErrServerPanic = errors.New("server panic")

// Recoverer allows recovering from panics.
type Recoverer struct {
	cfg  Config
	logs *zap.Logger
}

// NewRecoverer inits the recoverer.
func NewRecoverer(cfg Config, logs *zap.Logger) *Recoverer {
	return &Recoverer{cfg, logs}
}

func (r *Recoverer) handle(ctx context.Context, _ connect.Spec, _ http.Header, v any) error {
	clzap.Log(ctx, r.logs).Error("handling panic", zap.Any("recovered", v), zap.Stack("stack"))

	cerr := connect.NewError(connect.CodeInternal, ErrServerPanic)

	if !r.cfg.DisableStackTraceErrorDetails {
		addStackTraceDebugDetails(ctx, cerr)
	}

	return cerr
}

// addStackTraceDebugDetails add debug detail information to an connect error.
func addStackTraceDebugDetails(ctx context.Context, cerr *connect.Error) {
	debugInfo := &errdetails.DebugInfo{}
	for _, call := range stack.Trace().TrimRuntime() {
		debugInfo.StackEntries = append(debugInfo.GetStackEntries(), fmt.Sprintf("%+v", call))
	}

	if detail, derr := connect.NewErrorDetail(debugInfo); derr == nil {
		cerr.AddDetail(detail)
	} else {
		clzap.Log(ctx).Error("failed to init error detail", zap.Error(derr))
	}
}
