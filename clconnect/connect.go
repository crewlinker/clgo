// Package clconnect provides generic fx dependency for standard ConnectRPC services.
package clconnect

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/bufbuild/protovalidate-go"
	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the components.
type Config struct {
	// disables stack trace information in error details
	DisableStackTraceErrorDetails bool `env:"DISABLE_STACK_TRACE_ERROR_DETAILS"`
	// allows configuring the 'env' input field send to the policy, allows for configuring input
	// invariant to the environment
	AuthzPolicyEnvInput string `env:"AUTHZ_POLICY_ENV_INPUT,expand" envDefault:"{}"`
}

// ROTransacter is an interceptor that add read-only transactions to the context.
type ROTransacter interface {
	isROTransacter()
	connect.Interceptor
}

func (PgxROTransacter) isROTransacter()         {}
func (EntROTransactor[TX, MC]) isROTransacter() {}

// RWTransacter is an interceptor that add read-write transactions to the context.
type RWTransacter interface {
	isRWTransacter()
	connect.Interceptor
}

func (PgxRWTransacter) isRWTransacter()         {}
func (EntRWTransactor[TX, MC]) isRWTransacter() {}

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

	// middleware
	valr *validate.Interceptor,
	logr *Logger,
	rcvr *Recoverer,
	auth *Auth,
	rotx ROTransacter,
	rwtx RWTransacter,
) http.Handler {
	mux := http.NewServeMux()

	interceptors := connect.WithInterceptors(valr, logr, auth)
	recoverer := connect.WithRecover(rcvr.handle)

	rwopts := []connect.HandlerOption{interceptors, recoverer}
	if rwtx != nil {
		rwopts = append(rwopts, connect.WithInterceptors(rwtx))
	}

	rwp, rwh := rwc(rw, rwopts...)
	mux.Handle(rwp, rwh)

	roopts := []connect.HandlerOption{interceptors, recoverer}
	if rotx != nil {
		roopts = append(roopts, connect.WithInterceptors(rotx))
	}

	rop, roh := roc(ro, roopts...)
	mux.Handle(rop, roh)

	return mux
}

// to scope context keys.
type ctxKey string

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
		fx.Provide(fx.Annotate(New[RO, RW],
			// the transacters are optional, so we can use connect rpc without
			fx.ParamTags(``, ``, ``, ``, ``, ``, ``, ``, ``, ``, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"`+name+`"`))),
		// provide middleware constructors
		fx.Provide(protovalidate.New, NewRecoverer, NewLogger, NewAuth),
		// provide the validator interceptor
		fx.Provide(func(val *protovalidate.Validator) (*validate.Interceptor, error) {
			return validate.NewInterceptor(validate.WithValidator(val)) //nolint:wrapcheck
		}),
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
