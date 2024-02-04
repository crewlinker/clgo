// Package clmysql provides modules for interacting with MySQL.
package clmysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the package.
type Config struct {
	// ReadWriteHostname endpoint allows configuration of a endpoint that can read and write
	ReadOnlyDataSourceName string `env:"READ_ONLY_DATA_SOURCE_NAME" envDefault:"root:mysql@tcp(localhost:3306)/mysql"`
	// ReadWriteDataSourceName describes the data-source-name (DSN) for the read-write connection
	ReadWriteDataSourceName string `env:"READ_WRITE_DATA_SOURCE_NAME" envDefault:"root:mysql@tcp(localhost:3306)/mysql"`
}

// Migrater implements a migtration strategy that can be run before
// the database connection (pool) is setup.
type Migrater interface {
	Migrate(ctx context.Context) error
	Reset(ctx context.Context) error
}

// New inits a standard sql.DB with optional OTEL tracing and metrics. Any other dependency can optionally be
// provided as migrated to force it's lifecycle to be run before the database is connected. This is mostly useful to
// run migration logic (such as initializing the database).
func New(mycfg *mysql.Config, _ Migrater, trp trace.TracerProvider,
	mtp metric.MeterProvider,
) (*sql.DB, error) {
	connr, err := mysql.NewConnector(mycfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init connector: %w", err)
	}

	if trp != nil {
		attr := otelsql.WithAttributes(
			semconv.DBUserKey.String(mycfg.User),
			semconv.DBNameKey.String(mycfg.DBName),
			semconv.DBSystemMySQL)

		// trace sql
		db := otelsql.OpenDB(connr,
			otelsql.WithSpanOptions(otelsql.SpanOptions{Ping: true}),
			otelsql.WithTracerProvider(trp), attr)

		// add metrics
		if err := otelsql.RegisterDBStatsMetrics(db,
			otelsql.WithMeterProvider(mtp), attr); err != nil {
			return nil, fmt.Errorf("failed to register metrics: %w", err)
		}

		return db, nil
	}

	return sql.OpenDB(connr), nil
}

// moduleName for naming conventions.
const moduleName = "clmysql"

// Provide configures the DI for providng database connectivity.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide read/write configuration
		fx.Provide(fx.Annotate(NewReadOnlyConfig,
			fx.ResultTags(`name:"my_ro"`))),
		fx.Provide(fx.Annotate(NewReadWriteConfig,
			fx.ResultTags(`name:"my_rw"`))),

		// setup read-only *sql.DB stdlib connection pool
		fx.Provide(fx.Annotate(New,
			fx.ParamTags(`name:"my_ro"`, `optional:"true"`, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"my_ro"`),
			fx.OnStop(func(ctx context.Context, in struct {
				fx.In
				DB *sql.DB `name:"my_ro"`
			},
			) error {
				if err := in.DB.Close(); err != nil {
					return fmt.Errorf("failed to close: %w", err)
				}

				return nil
			}),
		)),
		// setup read-write *sql.DB connection
		fx.Provide(fx.Annotate(New,
			fx.ParamTags(`name:"my_rw"`, `optional:"true"`, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"my_rw"`),
			fx.OnStop(func(ctx context.Context, in struct {
				fx.In
				DB *sql.DB `name:"my_rw"`
			},
			) error {
				if err := in.DB.Close(); err != nil {
					return fmt.Errorf("failed to close: %w", err)
				}

				return nil
			}),
		)),
	)
}

// TestProvide configures the DI for a test environment.
func TestProvide() fx.Option {
	return fx.Options(Provide(),
		// we re-provide the read-write sql db as an unnamed *sql.DB and config because that is
		// what we usually want in tests.
		fx.Provide(fx.Annotate(func(rw *sql.DB) *sql.DB { return rw }, fx.ParamTags(`name:"my_rw"`))),
		fx.Provide(fx.Annotate(func(rw *mysql.Config) *mysql.Config { return rw }, fx.ParamTags(`name:"my_rw"`))),
	)
}
