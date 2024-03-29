// Package clredis provides reusable components for using Redis
package clredis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures Redis through the environment.
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
	// ClientName allows the application to indicate its name so connections can be more easily debugged
	ClientName string `env:"CLIENT_NAME" envDefault:"unknown"`
}

// NewOptions parses our environment config into options for the Redis client.
func NewOptions(cfg Config, logs *zap.Logger) (*redis.UniversalOptions, error) {
	opts := &redis.UniversalOptions{
		Addrs:    cfg.Addrs,
		Username: cfg.Username,
		Password: cfg.Password,

		// note: taken from: https://t.ly/cixv
		ReadOnly:       false,
		RouteRandomly:  false,
		RouteByLatency: false,

		ClientName: cfg.ClientName,
	}

	// unfortunately, the redis client only support a global logger, so
	// we set it during configuration
	redis.SetLogger(NewLogger(logs.Named("client")))

	// enable tls if available, potentially with server name checking
	if cfg.EnableTLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
			MinVersion:         tls.VersionTLS12,
		}

		if cfg.TLSServerName != "" {
			opts.TLSConfig.InsecureSkipVerify = false
			opts.TLSConfig.ServerName = cfg.TLSServerName
		}
	}

	return opts, nil
}

// New inits a universal client and instruments it if available.
func New(
	opts *redis.UniversalOptions, _ *zap.Logger, tp trace.TracerProvider, mtr metric.MeterProvider,
) (redis.UniversalClient, error) {
	ruc := redis.NewUniversalClient(opts)
	if tp != nil {
		if err := redisotel.InstrumentTracing(ruc, redisotel.WithTracerProvider(tp)); err != nil {
			return nil, fmt.Errorf("failed to instrument with tracing: %w", err)
		}
	}

	if mtr != nil {
		if err := redisotel.InstrumentMetrics(ruc, redisotel.WithMeterProvider(mtr)); err != nil {
			return nil, fmt.Errorf("failed to instrument with metrics: %w", err)
		}
	}

	return ruc, nil
}

// moduleName for naming conventions.
const moduleName = "clredis"

// Provide configures the DI for providng database connectivity.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// Init redis options from env variables
		fx.Provide(fx.Annotate(NewOptions)),
		// Init the client and ping on start, close on shutdown
		fx.Provide(fx.Annotate(New,
			fx.ParamTags(``, ``, `optional:"true"`, `optional:"true"`),
			fx.OnStart(func(ctx context.Context, red redis.UniversalClient) error {
				if err := red.Ping(ctx).Err(); err != nil {
					return fmt.Errorf("failed to ping redis: %w", err)
				}

				return nil
			}),
			fx.OnStop(func(ctx context.Context, red redis.UniversalClient) error {
				if err := red.Close(); err != nil {
					return fmt.Errorf("failed to close redis conn: %w", err)
				}

				return nil
			}))),
	)
}

// TestProvide configures the DI for a test environment.
func TestProvide() fx.Option { return fx.Options(Provide()) }
