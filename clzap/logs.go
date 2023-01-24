package clzap

import (
	"context"
	"io"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Fx is a convenient option that configures fx to use the zap logger.
func Fx() fx.Option {
	return fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: l.Named("fx")}
	})
}

// moduleName for naming conventions
const moduleName = "clzap"

// Prod logging module. It can be used as a fx Module in production binaries to provide
// high-performance structured logging.
var Prod = fx.Module(moduleName,
	// provide environment based configuration
	fx.Provide(fx.Annotate(parseConfig, fx.ParamTags(`optional:"true"`))),
	// allow environmental config to configure the level at which to log
	fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
	// provide the zapper, make sure everything is synced on shutdown
	fx.Provide(fx.Annotate(zap.New, fx.OnStop(func(ctx context.Context, l *zap.Logger) error { return l.Sync() }))),
	// provide depependencies to build the prod logger
	fx.Provide(zapcore.NewCore, zapcore.NewJSONEncoder, zap.NewProductionEncoderConfig),
	// allow environment to configure where logs are being synced to
	fx.Provide(func(cfg Config) (w zapcore.WriteSyncer, err error) {
		w, _, err = zap.Open(cfg.Outputs...)
		return
	}),
)

// newObservedAndColse outputs a tee logging core that writes to an observed underlying core and also writes
// console output to the configured writer.
func newObservedAndConsole(le zapcore.LevelEnabler, gw io.Writer) (c zapcore.Core, o *observer.ObservedLogs) {
	c, o = observer.New(le)
	c = zapcore.NewTee(c,
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(gw),
			le,
		))
	return
}

// Observed configures a logging module that allows for observing while also writing console output to
// a io.Writer that needs to be supplied.
var Observed = fx.Module(moduleName+"-observed",
	fx.Provide(fx.Annotate(parseConfig, fx.ParamTags(`optional:"true"`))),
	fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
	fx.Provide(newObservedAndConsole),
	fx.Provide(fx.Annotate(zap.New, fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
		return l.Sync()
	}))),
)

// Test is a convenient fx option setup that can easily be included in all tests. It observed the logs
// for assertion and writes console output to the GinkgoWriter so all logs can easily be inspected if
// tests fail.
var Test = fx.Options(Fx(),
	// in tests, always provide the ginkgo writer as the output writer so failing tests immediately show
	// the complete console output.
	fx.Supply(fx.Annotate(ginkgo.GinkgoWriter, fx.As(new(io.Writer)))),
	Observed)
