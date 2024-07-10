// Package clzap provides logging using the zap logging library
package clzap

import (
	"context"
	"fmt"
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

// Config configures the logging package.
type Config struct {
	// Level configures the minium logging level that will be captured.
	Level zapcore.Level `env:"LEVEL" envDefault:"info"`
	// Outputs configures the zap outputs that will be opened for logging.
	Outputs []string `env:"OUTPUTS" envDefault:"stderr"`
	// Configure the level at which fx logs are shown, default to debug
	FxLevel zapcore.Level `env:"FX_LEVEL" envDefault:"debug"`
	// By default it logs to lambda format, this
	DisableLambdaEncoding bool `env:"DISABLE_LAMBDA_ENCODING"`
	// Enables console encoding for more developer friendly logging output
	ConsoleEncoding bool `env:"CONSOLE_ENCODING"`
	// DevelopmentEncodingConfig enables encoding useful for developers.
	DevelopmentEncodingConfig bool `env:"DEVELOPMENT_ENCODING_CONFIG"`
}

// Fx is a convenient option that configures fx to use the zap logger.
func Fx() fx.Option {
	return fx.WithLogger(func(l *zap.Logger, cfg Config) fxevent.Logger {
		zl := &fxevent.ZapLogger{Logger: l.Named("fx")}
		zl.UseLogLevel(cfg.FxLevel)

		return zl
	})
}

// SecondaryCore can be provided in the DI to make the logger write to a second core. This is useful
// for writing logs to a observability system like Sentry.
type SecondaryCore = struct {
	zapcore.Core
	Name string
}

// newLogger can be used to create a logger from a single core or from two cores: writing to both.
func newLogger(zc zapcore.Core, sc *SecondaryCore, cfg Config) *zap.Logger {
	opts := []zap.Option{}

	if cfg.DevelopmentEncodingConfig {
		opts = append(opts, zap.Development())
	}

	l := zap.New(zc, opts...)
	if sc == nil {
		return l
	} else {
		l := zap.New(zapcore.NewTee(zc, sc.Core))
		l.Info("logger initialized with secondary core", zap.String("second_core_name", sc.Name))

		return l
	}
}

// newEncoder constructs the encoder based on the encoder config and our env config.
func newEncoder(cfg Config, ecfg zapcore.EncoderConfig) zapcore.Encoder {
	if cfg.ConsoleEncoding {
		return zapcore.NewConsoleEncoder(ecfg)
	}

	return zapcore.NewJSONEncoder(ecfg)
}

// newEncoderConfig constructs the encoder configuration.
func newEncoderConfig(cfg Config) zapcore.EncoderConfig {
	if cfg.DevelopmentEncodingConfig {
		return zap.NewDevelopmentEncoderConfig()
	}

	return zap.NewProductionEncoderConfig()
}

// moduleName for naming conventions.
const moduleName = "clzap"

// Provide logging module. It can be used as a fx Module in production binaries to provide
// high-performance structured logging.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// allow environmental config to configure the level at which to log
		fx.Provide(func(cfg Config) zapcore.LevelEnabler { return cfg.Level }),
		// provide the zapper, make sure everything is synced on shutdown
		fx.Provide(fx.Annotate(newLogger,
			fx.ParamTags(``, `optional:"true"`),
			fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
				_ = l.Sync() // ignore to support TTY: https://github.com/uber-go/zap/issues/880

				return nil
			}))),
		// provide dependencies to build the prod logger
		fx.Provide(zapcore.NewCore, newEncoder, newEncoderConfig),
		// customize the to fit AWS Lambda's application log format standard, as documented:
		// https://docs.aws.amazon.com/lambda/latest/dg/monitoring-cloudwatchlogs.html#monitoring-cloudwatchlogs-advanced
		fx.Decorate(func(cfg Config, ec zapcore.EncoderConfig) zapcore.EncoderConfig {
			if cfg.DisableLambdaEncoding {
				return ec
			}

			ec.MessageKey = "message"
			ec.TimeKey = "timestamp"
			ec.EncodeTime = zapcore.ISO8601TimeEncoder

			return ec
		}),
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

// TestProvide is a convenient fx option setup that can easily be included in all tests. It observed the logs
// for assertion and writes console output to the GinkgoWriter so all logs can easily be inspected if
// tests fail.
func TestProvide() fx.Option {
	return fx.Options(Fx(),
		// in tests, always provide the ginkgo writer as the output writer so failing tests immediately show
		// the complete console output.
		fx.Supply(fx.Annotate(ginkgo.GinkgoWriter, fx.As(new(io.Writer)))),
		Observed())
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
		fx.Provide(fx.Annotate(newLogger,
			fx.ParamTags(``, `optional:"true"`),
			fx.OnStop(func(ctx context.Context, l *zap.Logger) error {
				if err := l.Sync(); err != nil {
					return fmt.Errorf("failed to sync: %w", err)
				}

				return nil
			}))),
	)
}
