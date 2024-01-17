package clconnect

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clauthz"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
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

// AuthzInput encodes the full input into the authorization (AuthZ) policy system OPA.
// It should provide ALL data required to make authorization decisions. It should be fully
// serializable to JSON.
type AuthzInput struct {
	// OpenID token
	OpenID openid.Token `json:"open_id"`
	// Procedure encodes the full RPC procedure name. e.g: /acme.foo.v1.FooService/Bar
	Procedure string `json:"procedure"`
}

// intercept implements the actual authorization.
func (l Auth) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (resp connect.AnyResponse, err error) {
		bearer := strings.TrimSpace(strings.TrimPrefix(req.Header().Get("Authorization"), "Bearer"))
		token := openid.New() // anonymous token

		// authenticate non-anonymous token
		if bearer != "" {
			token, err = l.authn.AuthenticateJWT(ctx, []byte(bearer))
			if err != nil {
				l.logs.Info("failed to authenticate JWT",
					zap.String("raw_header", req.Header().Get("Authorization")),
					zap.String("bearer_token", bearer),
					zap.NamedError("auth_err", err))

				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
		}

		// authorization input
		input := &AuthzInput{
			OpenID:    token,
			Procedure: req.Spec().Procedure,
		}

		// authorize
		isAuthorized, err := l.authz.IsAuthorized(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error while authorizing: %w", err)
		} else if !isAuthorized {
			l.logs.Info("failed to authorize", zap.Any("token", token))

			return nil, connect.NewError(connect.CodePermissionDenied,
				fmt.Errorf("unauthorized, subject: '%s'", token.Subject()))
		}

		return next(ctx, req)
	})
}
