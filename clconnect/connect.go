// Package clconnect provides generic fx dependency for standard ConnectRPC services.
package clconnect

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the components.
type Config struct{}

// ConstructHandler defines the type for constructing a connectrpc service handler.
type ConstructHandler[SH any] func(svc SH, opts ...connect.HandlerOption) (string, http.Handler)

// ConstructClient is a funct signature that constructs a client.
type ConstructClient[SC any] func(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) SC

// New inits an http handler for the full RPC api.
func New[RO, RW any](
	cfg Config,
	logs *zap.Logger,
	ro RO, roc ConstructHandler[RO],
	rw RW, rwc ConstructHandler[RW],
) http.Handler {
	mux := http.NewServeMux()

	rwp, rwh := rwc(rw)
	mux.Handle(rwp, rwh)

	rop, roh := roc(ro)
	mux.Handle(rop, roh)

	return mux
}

// moduleName for naming conventions.
const moduleName = "clconnect"

// Provide the package components for the DI container.
func Provide[RO, RW any](name string) fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide as a named http handler
		fx.Provide(fx.Annotate(New[RO, RW], fx.ResultTags(`name:"`+name+`"`))),
	)
}

// TestProvide provides dependencies for testing.
func TestProvide[RO, RW, ROC, ROW any](name string) fx.Option {
	return fx.Options(
		Provide[RO, RW](name),

		// setup an test server for test clients to use
		fx.Provide(fx.Annotate(func(h http.Handler, lc fx.Lifecycle) *httptest.Server {
			s := httptest.NewServer(h)
			lc.Append(fx.StopHook(s.Close))

			return s
		}, fx.ParamTags(`name:"`+name+`"`))),
		// provide test clients for base rpc service
		fx.Provide(func(s *httptest.Server, scf ConstructClient[ROC]) ROC {
			return scf(http.DefaultClient, s.URL)
		}),
		fx.Provide(func(s *httptest.Server, scf ConstructClient[ROW]) ROW {
			return scf(http.DefaultClient, s.URL)
		}),
	)
}
