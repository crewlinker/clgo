package clory

import (
	"context"

	"connectrpc.com/connect"
)

// IsPublicRPCProcedure returns true if a request is done to public rpc method.
func (o Ory) IsPublicRPCMethod(spec connect.Spec) bool {
	return o.cfg.PublicRPCProcedures[spec.Procedure]
}

// Interceptor returns a ConnectRP interceptor that performs authentication (authn).
func (o Ory) Interceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			sess, err := o.Authenticate(ctx, req.Header().Get("cookie"), o.IsPublicRPCMethod(req.Spec()))
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			ctx = WithSession(ctx, sess)

			return next(ctx, req)
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
