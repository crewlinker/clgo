// Package claws provides the official AWS SDK (v2)
package claws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
		opts = append(opts, config.WithCredentialsProvider(aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.OverwriteAccessKeyID,
				cfg.OverwriteSecretAccessKey,
				cfg.OverwriteSessionToken,
			)),
		),
		)
	}

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

// Prod configures the DI for providng database connectivity.
func Prod() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide the actual aws config
		fx.Provide(fx.Annotate(New, fx.ParamTags(``, ``, `optional:"true"`, `optional:"true"`))),
	)
}

// DynamoEndpointDecorator will change the resolvers to set the dynamodb endpoint since this AWS supports a
// local version of Dynamo.
func DynamoEndpointDecorator(epurl string) func(c aws.Config) aws.Config {
	return func(acfg aws.Config) aws.Config {
		acfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(
			service, region string, options ...any,
		) (aws.Endpoint, error) {
			var aep aws.Endpoint
			switch service {
			case dynamodb.ServiceID:
				aep.URL = epurl

				return aep, nil
			default:
				return aep, &aws.EndpointNotFoundError{}
			}
		})

		return acfg
	}
}
