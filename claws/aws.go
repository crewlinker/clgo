package claws

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/caarlos0/env/v6"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures this package
type Config struct {
	// LoadConfigTimeout bounds the time given to config loading
	LoadConfigTimeout time.Duration `env:"LOAD_CONFIG_TIMEOUT" envDefault:"100ms"`
	// DynamoEndpoint allows configuring the dynamodb endpoint for testing because it supports a local version
	DynamoEndpoint *url.URL `env:"DYNAMO_ENDPOINT"`
}

// New initialize an AWS config to be used to create clients for individual aws services. We would like
// run this during fx lifecycle phase to provide it with a context because it can block. But too many
// dependencies would have to wait for it.
func New(
	cfg Config,
	logs *zap.Logger,
	epresolver aws.EndpointResolverWithOptions,
	tp trace.TracerProvider,
	pr propagation.TextMapPropagator,
) (acfg aws.Config, err error) {
	logs.Info("loading config", zap.Duration("timeout", cfg.LoadConfigTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), cfg.LoadConfigTimeout)
	defer cancel()

	if acfg, err = config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(epresolver)); err != nil {
		return acfg, fmt.Errorf("failed to load default config: %w", err)
	}

	// if we have a tracing available, we instrument the aws client
	if tp != nil {
		logs.Info("tracing provided, instrumenting aws client")
		otelaws.AppendMiddlewares(
			&acfg.APIOptions,
			otelaws.WithTracerProvider(tp),
			otelaws.WithTextMapPropagator(pr))
	}

	return acfg, nil
}

// moduleName for naming conventions
const moduleName = "claws"

// Prod configures the DI for providng database connectivity
var Prod = fx.Module(moduleName,
	// the incoming logger will be named after the module
	fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
	// provide the environment configuration
	fx.Provide(fx.Annotate(
		func(o env.Options) (c Config, err error) {
			o.Prefix = strings.ToUpper(moduleName) + "_"
			return c, env.Parse(&c, o)
		},
		fx.ParamTags(`optional:"true"`))),
	// provide the actual aws config
	fx.Provide(fx.Annotate(New, fx.ParamTags(``, ``, ``, `optional:"true"`, `optional:"true"`))),
	// provide endpoint resolver, can be used to ovewrite endpoints based on configuration
	fx.Provide(fx.Annotate(func(cfg Config) aws.EndpointResolverWithOptions {
		return aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (ep aws.Endpoint, err error) {
			switch {
			case service == dynamodb.ServiceID && cfg.DynamoEndpoint != nil:
				ep.URL = cfg.DynamoEndpoint.String()
				return ep, err
			default:
				return ep, &aws.EndpointNotFoundError{}
			}
		})
	})),
)
