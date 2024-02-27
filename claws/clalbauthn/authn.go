// Package clalbauthn functionality for dealing with requests that passed trough AWS ALB authentication action.
package clalbauthn

import (
	"context"
	"net/http"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures this package.
type Config struct {
	// If this is set the middleware will NOT return a 401 if the request does not have all the authn headers.
	AllowUnauthenticated bool `env:"ALLOW_UNAUTHENTICATED"`
}

// ctsKey provides a way to scope the context values to this package.
type ctxKey string

// AccessToken returns the x-amzn-oidc-accesstoken value, or an empty string.
func AccessToken(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey("access_token")).(string)

	return v
}

// Identity returns the x-amzn-oidc-identity value, or an empty string.
func Identity(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey("identity")).(string)

	return v
}

// ClaimData returns the x-amzn-oidc-data value, or an empty string.
func ClaimData(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey("claim_data")).(string)

	return v
}

// WithAccessToken returns a context with the token value.
func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxKey("access_token"), token)
}

// WithIdentity returns a context with the identity value.
func WithIdentity(ctx context.Context, identity string) context.Context {
	return context.WithValue(ctx, ctxKey("identity"), identity)
}

// WithClaimData returns a context with the claim data value.
func WithClaimData(ctx context.Context, data string) context.Context {
	return context.WithValue(ctx, ctxKey("claim_data"), data)
}

// Authentication provides middleware that adds the AWS ALB authentication claims
// in the request context.
type Authentication struct {
	cfg  Config
	errf func(w http.ResponseWriter, msg string, code int)
}

// Handler returns the middleware.
func (a Authentication) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accessToken, identity, data := req.Header.Get("x-amzn-oidc-accesstoken"),
			req.Header.Get("x-amzn-oidc-identity"),
			req.Header.Get("x-amzn-oidc-data")

		// block access to further handling if any of the alb headers are missing and it's not
		// explicitly allowed.
		if (accessToken == "" || identity == "" || data == "") && !a.cfg.AllowUnauthenticated {
			a.errf(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		ctx := WithAccessToken(req.Context(), accessToken)
		ctx = WithIdentity(ctx, identity)
		ctx = WithClaimData(ctx, data)

		// pass down the middleware chain.
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// UnauthenticatedHandler provides a handler for when the request is unauthenticated.
type UnauthenticatedHandler func(w http.ResponseWriter, msg string, code int)

// New inits the authentication middleware.
func New(cfg Config, unauthf UnauthenticatedHandler) *Authentication {
	return &Authentication{cfg: cfg, errf: unauthf}
}

// moduleName for naming conventions.
const moduleName = "clalbauthn"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the middleware
		fx.Provide(New),
		// provide http.Error as the default unauthenticated handler.
		fx.Supply(UnauthenticatedHandler(http.Error)),
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}
