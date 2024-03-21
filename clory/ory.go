// Package clory provides ory-powered auth functionality.
package clory

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clzap"
	"go.uber.org/fx"
	"go.uber.org/zap"

	orysdk "github.com/ory/client-go"
)

// Config configures this package.
type Config struct {
	// Ory endpoint.
	Endpoint *url.URL `env:"ENDPOINT" envDefault:"http://localhost:4000"`
	// PublicRPCProcedures configures the ConnectRPC methods that are plublic. For these procedures a special
	// "anonymous" session will be passed to other middleware.
	PublicRPCProcedures map[string]bool `env:"PUBLIC_RPC_PROCEDURES"`
}

// Ory auth module.
type Ory struct {
	cfg   Config
	logs  *zap.Logger
	front FrontendAPI
	perm  PermissionAPI
}

// FrontendAPI describe the external interface.
type FrontendAPI interface {
	ToSession(ctx context.Context) orysdk.FrontendAPIToSessionRequest
	ToSessionExecute(r orysdk.FrontendAPIToSessionRequest) (*orysdk.Session, *http.Response, error)
}

// PermissionAPI describe the external interface.
type PermissionAPI interface {
	CheckPermission(ctx context.Context) orysdk.PermissionAPICheckPermissionRequest
	CheckPermissionExecute(r orysdk.PermissionAPICheckPermissionRequest) (
		*orysdk.CheckPermissionResult, *http.Response, error)
}

// NewClientAPIs inits the raw Ory SDK API Client.
func NewClientAPIs(cfg Config) (FrontendAPI, PermissionAPI) {
	ocfg := orysdk.NewConfiguration()
	ocfg.Servers = []orysdk.ServerConfiguration{{URL: cfg.Endpoint.String()}}

	client := orysdk.NewAPIClient(ocfg)

	return client.FrontendAPI, client.PermissionAPI
}

// New inits the ory auth module.
func New(cfg Config, logs *zap.Logger, front FrontendAPI, perm PermissionAPI) *Ory {
	return &Ory{cfg: cfg, logs: logs, front: front, perm: perm}
}

// BrowserLoginURL returns the url for starting the login flow.
func (o Ory) BrowserLoginURL() *url.URL {
	ep := *o.cfg.Endpoint
	ep.Path = "/self-service/login/browser"

	return &ep
}

// ErrUnauthenticated defines an error for failing to authenticate.
var ErrUnauthenticated = errors.New("unauthenticated")

// AnonymousSessionID declares the special session id for anonymous session.
const AnonymousSessionID = "00000000-0000-4000-8000-000000000001"

// AnonymousIdentityID declares the special identity id for anonymous users.
const AnonymousIdentityID = "00000000-0000-4000-8000-000000000002"

// AnonymousSession is returned from authentication calls when they fail but anonymous
// access is allowed.
var AnonymousSession = &orysdk.Session{
	Id:       AnonymousSessionID,   // anonymous sessions have this fixed session id
	Active:   orysdk.PtrBool(true), // anonymous sessions are always active
	Identity: &orysdk.Identity{Id: AnonymousIdentityID},
}

// onUnauthenticated provides centralized logic for the case that a user is unauthenticated.
func (o Ory) onUnauthenticated(
	ctx context.Context,
	tStart time.Time,
	unerlyingErr error,
	internalReason string,
	allowAnonymous bool,
) (*orysdk.Session, error) {
	if allowAnonymous {
		sess := AnonymousSession

		clzap.Log(ctx, o.logs).Info("authenticated as anonymous",
			zap.String("reason", internalReason),
			zap.Any("session", sess),
			zap.Error(unerlyingErr), zap.Duration("rtt", time.Since(tStart)))

		return sess, nil
	}

	clzap.Log(ctx, o.logs).Info("unauthenticated",
		zap.String("reason", internalReason),
		zap.Error(unerlyingErr), zap.Duration("rtt", time.Since(tStart)))

	return nil, ErrUnauthenticated
}

// Authenticate implements the core authentication logic. The function that actually interacts
// with Ory is passed in for easier testing.
func (o Ory) Authenticate(ctx context.Context, cookie string, allowAnonymous bool) (*orysdk.Session, error) {
	req, tStart := o.front.ToSession(ctx), time.Now()
	req = req.Cookie(cookie)

	sess, _, err := o.front.ToSessionExecute(req) //nolint:bodyclose // closed by library
	if err != nil {
		return o.onUnauthenticated(ctx, tStart, err, "error from Ory", allowAnonymous)
	} else {
		if sess == nil || sess.Active == nil {
			return o.onUnauthenticated(ctx, tStart, nil, "session or active state is nil", allowAnonymous)
		}

		if !*sess.Active {
			return o.onUnauthenticated(ctx, tStart, nil, "session is inactive", allowAnonymous)
		} else {
			clzap.Log(ctx, o.logs).Info("authenticated",
				zap.Any("session", sess),
				zap.Duration("rtt", time.Since(tStart)))

			return sess, nil // OK
		}
	}
}

// Private middleware will try to fetch a Ory session for the request or else return
// a 401 Unauthorized response.
func (o Ory) Private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		sess, err := o.Authenticate(ctx, req.Header.Get("Cookie"), false)
		if err != nil {
			resp.Header().Set("X-Browser-Login-URL", o.BrowserLoginURL().String())
			http.Error(resp, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

			return
		}

		ctx = WithSession(ctx, sess)

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
