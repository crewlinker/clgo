// Package clredis provides reusable components for using Redis
package clredis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures Redis through the environment
type Config struct {
	// Node addresses provide to the cluster client
	Addrs []string `env:"ADDRS" envDefault:"localhost:6379"`
	// Username for authenticated configuration
	Username string `env:"USERNAME"`
	// Password for authenticated configuration
	Password string `env:"PASSWORD"`
	// Enable tls
	EnableTLS bool `env:"ENABLE_TLS" envDefault:"false"`
	// Setting this enables TLS connections
	TLSServerName string `env:"TLS_SERVER_NAME"`
}

// NewOptions parses our environment config into options for the Redis client
func NewOptions(cfg Config, logs *zap.Logger) (*redis.UniversalOptions, error) {
	opts := &redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		Username: cfg.Username,
		Password: cfg.Password,

		// note: taken from https://stackoverflow.com/questions/73907312/i-want-to-connect-to-elasticcache-for-redis-in-which-cluster-mode-is-enabled-i
		ReadOnly:       false,
		RouteRandomly:  false,
		RouteByLatency: false,
	}

	// unfortunately, the redis client only support a global logger, so
	// we set it during configuration
	redis.SetLogger(NewLogger(logs.Named("client")))

	// enable tls if available, potentially with server name checking
	if cfg.EnableTLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
		}

		if cfg.TLSServerName != "" {
			opts.TLSConfig.InsecureSkipVerify = false
			opts.TLSConfig.ServerName = cfg.TLSServerName
		}
	}

	return opts, nil
}

// New inits a univeral client and instruments it if available
func New(opts *redis.UniversalOptions, logs *zap.Logger, tp trace.TracerProvider, mp metric.MeterProvider) (redis.UniversalClient, error) {
	ruc := redis.NewUniversalClient(opts)
	if tp != nil {
		if err := redisotel.InstrumentTracing(ruc, redisotel.WithTracerProvider(tp)); err != nil {
			return nil, fmt.Errorf("failed to instrument with tracing: %w", err)
		}
	}

	if mp != nil {
		if err := redisotel.InstrumentMetrics(ruc, redisotel.WithMeterProvider(mp)); err != nil {
			return nil, fmt.Errorf("failed to instrument with metrics: %w", err)
		}
	}

	return ruc, nil
}

// moduleName for naming conventions
const moduleName = "clredis"

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
	// Init redis options from env variables
	fx.Provide(fx.Annotate(NewOptions)),
	// Init the client and ping on start, close on shutdown
	fx.Provide(fx.Annotate(New,
		fx.ParamTags(``, ``, `optional:"true"`, `optional:"true"`),
		fx.OnStart(func(ctx context.Context, red redis.UniversalClient) error {
			return red.Ping(ctx).Err()
		}),
		fx.OnStop(func(ctx context.Context, red redis.UniversalClient) error {
			return red.Close()
		}))),
)

// Test configures the DI for a test environment
var Test = fx.Options(Prod)
