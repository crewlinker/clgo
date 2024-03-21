package clconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clauthz"
	"github.com/lestrrat-go/jwx/v2/jwt/openid"
	"go.uber.org/zap"
)

// JWTOPAAuth provides authn and authz as an injector.
type JWTOPAAuth struct {
	cfg   Config
	logs  *zap.Logger
	authn *clauthn.Authn
	authz *clauthz.Authz

	envInput map[string]any

	connect.Interceptor
}

// NewJWTOPAAuth inits an interceptor that uses JWT for Authn and OPA for Authz.
func NewJWTOPAAuth(
	cfg Config, logs *zap.Logger, authn *clauthn.Authn, authz *clauthz.Authz,
) (lgr *JWTOPAAuth, err error) {
	lgr = &JWTOPAAuth{
		cfg:      cfg,
		logs:     logs.Named("jwt_opa_auth"),
		authn:    authn,
		authz:    authz,
		envInput: map[string]any{},
	}

	if err = json.Unmarshal([]byte(cfg.AuthzPolicyEnvInput), &lgr.envInput); err != nil {
		return nil, fmt.Errorf("failed to parse authz policy env input `%s`: %w", cfg.AuthzPolicyEnvInput, err)
	}

	lgr.Interceptor = connect.UnaryInterceptorFunc(lgr.intercept)

	return lgr, nil
}

// WithIdentity returns a context with the openid token.
func WithIdentity(ctx context.Context, tok openid.Token) context.Context {
	return context.WithValue(ctx, ctxKey("openid_token"), tok)
}

// IdentityFromContext returns the identity from context as an OpenID token. If there is no
// token in the context it returns an empty (anonymous) openid token.
func IdentityFromContext(ctx context.Context) openid.Token {
	v, ok := ctx.Value(ctxKey("openid_token")).(openid.Token)
	if !ok || v == nil {
		v = openid.New() // anonymous token
	}

	return v
}

// AuthzInput encodes the full input into the authorization (AuthZ) policy system OPA.
// It should provide ALL data required to make authorization decisions. It should be fully
// serializable to JSON.
type AuthzInput struct {
	// Input from the process environment
	Env map[string]any `json:"env"`
	// OpenID token as claims
	Claims openid.Token `json:"claims"`
	// Procedure encodes the full RPC procedure name. e.g: /acme.foo.v1.FooService/Bar
	Procedure string `json:"procedure"`
}

// intercept implements the actual authorization.
func (l JWTOPAAuth) intercept(next connect.UnaryFunc) connect.UnaryFunc {
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
			Env:       l.envInput,
			Claims:    token,
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

		ctx = WithIdentity(ctx, token)

		return next(ctx, req)
	})
}
