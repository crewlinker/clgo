// Package clsentry provides Sentry error reporting.
package clsentry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clzap"
	sentry "github.com/getsentry/sentry-go"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
)

// Config configures.
type Config struct {
	// Sentry DNS to send data to.
	DNS string `env:"DNS"`
	// TracesSampleRate is the rate at which traces are captures for Sentry.
	TracesSampleRate float64 `env:"TRACES_SAMPLE_RATE" envDefault:"1.0"`
	// DefaultFlushTimeout is the default timeout for flushing to Sentry.
	DefaultFlushTimeout time.Duration `env:"DEFAULT_FLUSH_TIMEOUT" envDefault:"2s"`
	// AttachStacktrace is whether to attach stacktrace to pure capture message calls.
	AttachStacktrace bool `env:"ATTACH_STACKTRACE" envDefault:"true"`
	// ZapSentryLevel is the level at which zap will send messages to Sentry.
	ZapSentryLevel zapcore.Level `env:"ZAP_SENTRY_LEVEL" envDefault:"warn"`
	// ZapSentryBreadcrumbLevel is the level at which zap will send breadcrumbs to Sentry.
	ZapSentryBreadcrumbLevel zapcore.Level `env:"ZAP_SENTRY_BREADCRUMB_LEVEL" envDefault:"info"`
}

// NewZapSentry create a secondary zap core. The clzap package will automatically pick it up and make
// sure all logs are sent to it as well.
func NewZapSentry(cfg Config, hub *sentry.Hub, client *sentry.Client) (*clzap.SecondaryCore, error) {
	zsCfg := zapsentry.Configuration{
		Hub:               hub,
		Level:             cfg.ZapSentryLevel,
		EnableBreadcrumbs: true,
		BreadcrumbLevel:   cfg.ZapSentryBreadcrumbLevel,
	}

	core, err := zapsentry.NewCore(zsCfg, zapsentry.NewSentryClientFromClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create zap-sentry core: %w", err)
	}

	return &clzap.SecondaryCore{core}, nil
}

// newOptions creates a new sentry client options.
func newOptions(cfg Config, binfo clbuildinfo.Info) sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:              cfg.DNS,
		TracesSampleRate: cfg.TracesSampleRate,
		AttachStacktrace: cfg.AttachStacktrace,

		Release: binfo.Version(),
	}
}

// newHub inits a new sentry hub.
func newHub(client *sentry.Client, scope *sentry.Scope) *sentry.Hub {
	return sentry.NewHub(client, scope)
}

// delayOnFxError is a fx error handler hook that delays fx when an error occurs. This will allow the
// sentry hub to flush the error before the application exits.
type delayOnFxError struct{}

// FxErrorShutdownDelay is the delay to wait before shutting down the application after an error. It allows flushing
// of the sentry hub to complete.
var FxErrorShutdownDelay = time.Second * 2

func (h delayOnFxError) HandleError(_ error) {
	time.Sleep(FxErrorShutdownDelay)
}

// ErrFlushFailed is returned when the flush fails during fx shutdown.
var ErrFlushFailed = errors.New("failed to flush sentry hub")

// moduleName for naming conventions.
const moduleName = "clsentry"

// Provide configures the DI for providng rpc.
func Provide() fx.Option {
	var cfg Config

	return fx.Module(moduleName,
		fx.Provide(sentry.NewClient, sentry.NewScope, newOptions, NewZapSentry),
		fx.Populate(&cfg),

		fx.ErrorHook(delayOnFxError{}),

		// provide the sentry hub and flush on shutdown.
		fx.Provide(fx.Annotate(newHub, fx.OnStop(func(ctx context.Context, hub *sentry.Hub) error {
			dl, ok := ctx.Deadline()
			if !ok {
				dl = time.Now().Add(cfg.DefaultFlushTimeout)
			}

			if !hub.Flush(time.Until(dl)) {
				return fmt.Errorf("%w", ErrFlushFailed)
			}

			return nil
		}))),

		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
	)
}
