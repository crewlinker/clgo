// Package clzap provides logging using the zap logging library
package clzap

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Config configures the logging package.
type Config struct {
	// Level configures the minium logging level that will be captured.
	Level zapcore.Level `env:"LEVEL" envDefault:"info"`
	// Outputs configures the zap outputs that will be opened for logging.
	Outputs []string `env:"OUTPUTS" envDefault:"stderr"`
}

// Fx is a convenient option that configures fx to use the zap logger.
func Fx() fx.Option {
	return fx.WithLogger(func(l *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: l.Named("fx")}
	})
}

// moduleName for naming conventions.
const moduleName = "clzap"

// Prod logging module. It can be used as a fx Module in production binaries to provide
// high-performance structured logging.
func Prod() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// allow environmental config to configure the level at which to log
		fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
		// provide the zapper, make sure everything is synced on shutdown
		fx.Provide(fx.Annotate(zap.New, fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
			_ = l.Sync() // ignore to support TTY: https://github.com/uber-go/zap/issues/880

			return nil
		}))),
		// provide dependencies to build the prod logger
		fx.Provide(zapcore.NewCore, zapcore.NewJSONEncoder, zap.NewProductionEncoderConfig),
		// allow environment to configure where logs are being synced to
		fx.Provide(func(cfg Config) (zapcore.WriteSyncer, error) {
			sync, _, err := zap.Open(cfg.Outputs...)
			if err != nil {
				return nil, fmt.Errorf("failed to zap-open: %w", err)
			}

			return sync, nil
		}),
	)
}

// newObservedAndColse outputs a tee logging core that writes to an observed underlying core and also writes
// console output to the configured writer.
func newObservedAndConsole(lvl zapcore.LevelEnabler, gw io.Writer) (zapcore.Core, *observer.ObservedLogs) {
	core, obs := observer.New(lvl)
	core = zapcore.NewTee(core,
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(gw),
			lvl,
		))

	return core, obs
}

// Observed configures a logging module that allows for observing while also writing console output to
// a io.Writer that needs to be supplied.
func Observed() fx.Option {
	return fx.Module(moduleName+"-observed",
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
		fx.Provide(newObservedAndConsole),
		fx.Provide(fx.Annotate(zap.New, fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
			if err := l.Sync(); err != nil {
				return fmt.Errorf("failed to sync: %w", err)
			}

			return nil
		}))),
	)
}
