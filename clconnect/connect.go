// Package clconnect provides generic fx dependency for standard ConnectRPC services.
package clconnect

import (
	"fmt"
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

	// PublicRPCProcedures configures the ConnectRPC methods that are plublic. For these procedures a special
	// "anonymous" session will be passed to other middleware.
	PublicRPCProcedures map[string]bool `env:"PUBLIC_RPC_PROCEDURES"`
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
	rcvr *Recoverer,

	// required middleware (interceptors)
	logr *Logger,
	valr *validate.Interceptor,

	// optional middleware (interceptors)
	roTx ROTransacter, // optional
	rwTx RWTransacter, // optional
	joAuth *JWTOPAAuth, // optional
	oryAuth *OryAuth, // optional
) http.Handler {
	mux := http.NewServeMux()

	// base interceptors
	baseIntercepts := []connect.Interceptor{valr, logr}

	// optional injectors, check for nil
	{
		if joAuth != nil {
			baseIntercepts = append(baseIntercepts, joAuth)
		}

		if oryAuth != nil {
			baseIntercepts = append(baseIntercepts, oryAuth)
		}
	}

	// base options
	interceptors := connect.WithInterceptors(baseIntercepts...)
	recoverer := connect.WithRecover(rcvr.handle)

	// setup read-write specific options (interceptors)
	{
		rwopts := []connect.HandlerOption{interceptors, recoverer}
		if rwTx != nil {
			rwopts = append(rwopts, connect.WithInterceptors(rwTx))
		}

		rwp, rwh := rwc(rw, rwopts...)
		mux.Handle(rwp, rwh)
	}

	// setup read-only specific options (interceptors)
	{
		roopts := []connect.HandlerOption{interceptors, recoverer}
		if roTx != nil {
			roopts = append(roopts, connect.WithInterceptors(roTx))
		}

		rop, roh := roc(ro, roopts...)
		mux.Handle(rop, roh)
	}

	// finally, serve a "not found" handler that renders the most minimal error according to the spec
	// https://connectrpc.com/docs/protocol/#error-end-stream
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"code": "unimplemented"}`)
	})

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
			fx.ParamTags(``, ``, ``, ``, ``, ``, ``, ``, ``,
				`optional:"true"`, `optional:"true"`, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"`+name+`"`))),
		// provide mandatory middleware constructors
		fx.Provide(protovalidate.New, NewRecoverer, NewLogger),
		// provide the validator interceptor
		fx.Provide(func(val *protovalidate.Validator) (*validate.Interceptor, error) {
			return validate.NewInterceptor(validate.WithValidator(val))
		}),
	)
}

// TestMiddleware can be provided in tests to wrap the test http.Handler with middleware.
type TestMiddleware func(next http.Handler) http.Handler

// TestProvide provides dependencies for testing.
func TestProvide[RO, RW, ROC, ROW any](name string) fx.Option {
	return fx.Options(
		Provide[RO, RW](name),

		// setup an test server for test clients to use
		fx.Provide(fx.Annotate(func(h http.Handler, lc fx.Lifecycle, tmwr TestMiddleware) *httptest.Server {
			if tmwr != nil {
				h = tmwr(h) // wrap if test middleware is provided
			}

			s := httptest.NewServer(h)
			lc.Append(fx.StopHook(s.Close))

			return s
		}, fx.ParamTags(`name:"`+name+`"`, ``, `optional:"true"`))),
		// provide test clients for base rpc service
		fx.Provide(func(s *httptest.Server, scf ConstructClient[ROC]) ROC {
			return scf(http.DefaultClient, s.URL)
		}),
		fx.Provide(func(s *httptest.Server, scf ConstructClient[ROW]) ROW {
			return scf(http.DefaultClient, s.URL)
		}),
	)
}
