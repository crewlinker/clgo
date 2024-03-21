package clconnect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clory"
	"github.com/crewlinker/clgo/clzap"
	orysdk "github.com/ory/client-go"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Ory interface provides our interface onto ory.
type Ory interface {
	// Authenticate implements the authentication logic.
	Authenticate(ctx context.Context, cookie string, allowAnonymous bool) (*orysdk.Session, error)
}

// OryAuth provides authn and authz as an injector.
type OryAuth struct {
	cfg  Config
	logs *zap.Logger
	ory  Ory
	connect.Interceptor
}

// NewOryAuth inits an interceptor that uses JWT for Authn and OPA for Authz.
func NewOryAuth(
	cfg Config, logs *zap.Logger, ory Ory,
) (inj *OryAuth, err error) {
	inj = &OryAuth{
		cfg:  cfg,
		logs: logs.Named("ory_auth"),
		ory:  ory,
	}

	inj.Interceptor = connect.UnaryInterceptorFunc(inj.intercept)

	logs.Info("ory auth initialized", zap.Strings("public_rpc_procedures", lo.Keys(cfg.PublicRPCProcedures)))

	return inj, nil
}

// IsPublicRPCProcedure returns true if a request is done to public rpc method.
func (l OryAuth) IsPublicRPCMethod(spec connect.Spec) bool {
	return l.cfg.PublicRPCProcedures[spec.Procedure]
}

// intercept implements the actual authorization.
func (l OryAuth) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (resp connect.AnyResponse, err error) {
		clzap.Log(ctx, l.logs).Debug("auth interceptor started",
			zap.Any("headers", req.Header()),
			zap.String("http_method", req.HTTPMethod()),
			zap.Stringer("spec_idempotency_level", req.Spec().IdempotencyLevel),
			zap.String("spec_procedure", req.Spec().Procedure))

		sess, err := l.ory.Authenticate(ctx, req.Header().Get("cookie"), l.IsPublicRPCMethod(req.Spec()))
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, err)
		}

		ctx = clory.WithSession(ctx, sess)

		return next(ctx, req)
	})
}

// ProvideOryAuth provides injector for ory-based auth.
func ProvideOryAuth() fx.Option {
	return fx.Options(
		fx.Provide(NewOryAuth),
		fx.Provide(func(o *clory.Ory) Ory { return o }),
	)
}
