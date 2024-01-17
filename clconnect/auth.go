package clconnect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clauthz"
	"go.uber.org/zap"
)

// Auth provides authn and authz as an injector.
type Auth struct {
	cfg   Config
	logs  *zap.Logger
	authn *clauthn.Authn
	authz *clauthz.Authz

	connect.Interceptor
}

// NewAuth inits the logger.
func NewAuth(cfg Config, logs *zap.Logger, authn *clauthn.Authn, authz *clauthz.Authz) *Auth {
	lgr := &Auth{cfg: cfg, logs: logs.Named("auth"), authn: authn, authz: authz}
	lgr.Interceptor = connect.UnaryInterceptorFunc(lgr.intercept)

	return lgr
}

func (l Auth) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return next(ctx, req)
	})
}
