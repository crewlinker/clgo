// Package clzap provides logging using the zap logging library
package clzap

import (
	"context"
	"io"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/onsi/ginkgo/v2"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Config configures the logging package
type Config struct {
	// Level configures the the minium logging level that will be captured
	Level zapcore.Level `env:"LEVEL" envDefault:"info"`
	// Outputs configures the zap outputs that will be opened for logging
	Outputs []string `env:"OUTPUTS" envDefault:"stderr"`
}

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
	// provide the environment configuration
	clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
	// allow environmental config to configure the level at which to log
	fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
	// provide the zapper, make sure everything is synced on shutdown
	fx.Provide(fx.Annotate(zap.New, fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
		l.Sync() // ignore to support TTY: https://github.com/uber-go/zap/issues/880
		return nil
	}))),
	// provide dependencies to build the prod logger
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
	// provide the environment configuration
	clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
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
