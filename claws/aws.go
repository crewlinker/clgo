// Package claws provides the official AWS SDK (v2)
package claws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go/logging"
	"github.com/crewlinker/clgo/clconfig"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures this package.
type Config struct {
	// LoadConfigTimeout bounds the time given to config loading
	LoadConfigTimeout time.Duration `env:"LOAD_CONFIG_TIMEOUT" envDefault:"100ms"`
	// OverwriteAccessKeyID can be set to overwrite regular credentials loading chain and just a static key/secret
	OverwriteAccessKeyID string `env:"OVERWRITE_ACCESS_KEY_ID"`
	// If OverwriteAccessKeyID this wil be used as the secret
	OverwriteSecretAccessKey string `env:"OVERWRITE_SECRET_ACCESS_KEY"`
	// If OverwriteAccessKeyID this wil be used as the session token
	OverwriteSessionToken string `env:"OVERWRITE_SESSION_TOKEN"`

	// enable logging of SDK retries
	LogRetries bool `env:"LOG_RETRIES" envDefault:"false"`
	// enable logging of SDK request
	LogRequest bool `env:"LOG_REQUEST" envDefault:"false"`
	// enable logging of SDK responses
	LogResponse bool `env:"LOG_RESPONSE" envDefault:"false"`
}

// logger implements AWS logging.Logger using zap.
type Logger struct {
	warns  *log.Logger
	debugs *log.Logger
}

// Logf implements logging.Logger interface.
func (l Logger) Logf(classification logging.Classification, format string, v ...interface{}) {
	switch classification {
	case logging.Warn:
		l.warns.Printf(format, v...)
	case logging.Debug:
		fallthrough
	default:
		l.debugs.Printf(format, v...)
	}
}

// NewLogger creates a new logger for AWS SDK.
func NewLogger(logs *zap.Logger) *Logger {
	l := &Logger{}
	l.warns, _ = zap.NewStdLogAt(logs.Named("sdk"), zap.WarnLevel)
	l.debugs, _ = zap.NewStdLogAt(logs.Named("sdk"), zap.DebugLevel)

	return l
}

// New initialize an AWS config to be used to create clients for individual aws services. We would like
// run this during fx lifecycle phase to provide it with a context because it can block. But too many
// dependencies would have to wait for it.
func New(
	cfg Config,
	logs *zap.Logger,
	trp trace.TracerProvider,
	txtp propagation.TextMapPropagator,
) (aws.Config, error) {
	logs.Info("loading config", zap.Duration("timeout", cfg.LoadConfigTimeout))

	var (
		acfg aws.Config
		err  error
	)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.LoadConfigTimeout)
	defer cancel()

	opts := []func(*config.LoadOptions) error{}
	if cfg.OverwriteAccessKeyID != "" {
		opts = append(opts, config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					cfg.OverwriteAccessKeyID,
					cfg.OverwriteSecretAccessKey,
					cfg.OverwriteSessionToken,
				)),
		))
	}

	logMode := aws.LogDeprecatedUsage
	if cfg.LogRetries {
		logMode |= aws.LogRetries
	}

	if cfg.LogRequest {
		logMode |= aws.LogRequest
	}

	if cfg.LogResponse {
		logMode |= aws.LogResponse
	}

	opts = append(opts,
		config.WithLogger(NewLogger(logs)),
		config.WithClientLogMode(logMode))

	if acfg, err = config.LoadDefaultConfig(ctx, opts...); err != nil {
		return acfg, fmt.Errorf("failed to load default config: %w", err)
	}

	// if we have a tracing available, we instrument the aws client
	if trp != nil {
		logs.Info("tracing provided, instrumenting aws client")
		otelaws.AppendMiddlewares(
			&acfg.APIOptions,
			otelaws.WithTracerProvider(trp),
			otelaws.WithTextMapPropagator(txtp))
	}

	return acfg, nil
}

// moduleName for naming conventions.
const moduleName = "claws"

// Provide configures the DI for providng database connectivity.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the actual aws config
		fx.Provide(fx.Annotate(New, fx.ParamTags(``, ``, `optional:"true"`, `optional:"true"`))),
	)
}
