// Package clory provides ory-powered auth functionality.
package clory

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"

	orysdk "github.com/ory/client-go"
)

// Config configures this package.
type Config struct {
	// Ory endpoint.
	Endpoint *url.URL `env:"ENDPOINT" envDefault:"http://localhost:4000"`
}

// Ory auth module.
type Ory struct {
	cfg   Config
	logs  *zap.Logger
	front FrontendAPI
}

// FrontendAPI describe the external interface.
type FrontendAPI interface {
	ToSession(ctx context.Context) orysdk.FrontendAPIToSessionRequest
	ToSessionExecute(r orysdk.FrontendAPIToSessionRequest) (*orysdk.Session, *http.Response, error)
}

// NewClientAPIs inits the raw Ory SDK API Client.
func NewClientAPIs(cfg Config) FrontendAPI {
	ocfg := orysdk.NewConfiguration()
	ocfg.Servers = []orysdk.ServerConfiguration{{URL: cfg.Endpoint.String()}}

	client := orysdk.NewAPIClient(ocfg)

	return client.FrontendAPI
}

// New inits the ory auth module.
func New(cfg Config, logs *zap.Logger, front FrontendAPI) *Ory {
	return &Ory{cfg: cfg, logs: logs, front: front}
}

// BrowserLoginURL returns the url for starting the login flow.
func (o Ory) BrowserLoginURL() *url.URL {
	ep := *o.cfg.Endpoint
	ep.Path = "/self-service/login/browser"

	return &ep
}

// Private middleware will try to fetch a Ory session for the request or else return
// a 401 Unauthorized response.
func (o Ory) Private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		inp, t0 := o.front.ToSession(req.Context()), time.Now()
		inp = inp.Cookie(req.Header.Get("Cookie"))

		sess, _, err := o.front.ToSessionExecute(inp) //nolint:bodyclose // closed by library
		if (err != nil && sess == nil) || (err == nil && !*sess.Active) {
			o.logs.Info("unauthenticated", zap.Error(err), zap.Any("session", sess), zap.Duration("rtt", time.Since(t0)))

			resp.Header().Set("X-Browser-Login-URL", o.BrowserLoginURL().String())
			http.Error(resp, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		o.logs.Info("authenticated", zap.Any("session", sess), zap.Duration("rtt", time.Since(t0)))

		ctx := WithSession(req.Context(), sess)

		next.ServeHTTP(resp, req.WithContext(ctx))
	})
}

// Session returns the session stored in the context, or nil if it isn't.
func Session(ctx context.Context) *orysdk.Session {
	sess, _ := ctx.Value(ctxKey("session")).(*orysdk.Session)

	return sess
}

// WithSession adds the session to the middleware.
func WithSession(ctx context.Context, sess *orysdk.Session) context.Context {
	return context.WithValue(ctx, ctxKey("session"), sess)
}

// ctxKey isolates context values.
type ctxKey string

// moduleName for naming conventions.
const moduleName = "clory"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the http handler.
		fx.Provide(NewClientAPIs, New),
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	)
}
