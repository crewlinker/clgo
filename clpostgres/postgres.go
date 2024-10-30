// Package clpostgres provides re-usable components for using PostgreSQL
package clpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Migrater implements a migtration strategy that can be run before
// the database connection (pool) is setup.
type Migrater interface {
	Migrate(ctx context.Context) error
	Reset(ctx context.Context) error
}

// NewPool inits a raw pgx postgres connection pool. Migrater is specified as a dependency so it
// force it's lifecycle hooks to be run.
func NewPool(pcfg *pgxpool.Config, cfg Config, _ Migrater) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.PoolConnectionTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init pool with config: %w", err)
	}

	return pool, nil
}

// New inits a stdlib sql connection. Any other dependency can optionally be provided as migrated
// to force it's lifecycle to be run before the database is connected. This is mostly useful to
// run migration logic (such as initializing the database).
func New(
	pcfg *pgxpool.Config,
	_ Migrater,
	trp trace.TracerProvider,
	mtp metric.MeterProvider,
) (*sql.DB, error) {
	openopts := []stdlib.OptionOpenDB{}
	if pcfg.BeforeConnect != nil {
		openopts = append(openopts, stdlib.OptionBeforeConnect(pcfg.BeforeConnect)) // if set, for IAM auth
	}

	connr := stdlib.GetConnector(*pcfg.ConnConfig, openopts...)

	if trp != nil {
		attr := otelsql.WithAttributes(
			semconv.DBUserKey.String(pcfg.ConnConfig.User),
			semconv.DBNameKey.String(pcfg.ConnConfig.Database),
			semconv.DBSystemPostgreSQL)

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
const moduleName = "clpostgres"

// Provide configures the DI for providng database connectivity.
func Provide() fx.Option {
	return fx.Module(moduleName,
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),
		// the incoming logger will be named after the module
		fx.Decorate(func(l *zap.Logger) *zap.Logger { return l.Named(moduleName) }),
		// provide read/write configuration
		fx.Provide(fx.Annotate(NewReadOnlyConfig,
			fx.ParamTags(``, ``, `optional:"true"`), fx.ResultTags(`name:"ro"`))),
		fx.Provide(fx.Annotate(NewReadWriteConfig,
			fx.ParamTags(``, ``, `optional:"true"`), fx.ResultTags(`name:"rw"`))),
		// setup read-only *pgxpool.Pool connection
		fx.Provide(fx.Annotate(NewPool,
			fx.ParamTags(`name:"ro"`, ``, `optional:"true"`),
			fx.ResultTags(`name:"ro"`),
			fx.OnStop(func(in struct {
				fx.In
				DB *pgxpool.Pool `name:"ro"`
			},
			) {
				in.DB.Close()
			}),
		)),
		// setup read-write *pgxpool.Pool connection
		fx.Provide(fx.Annotate(NewPool,
			fx.ParamTags(`name:"rw"`, ``, `optional:"true"`),
			fx.ResultTags(`name:"rw"`),
			fx.OnStop(func(in struct {
				fx.In
				DB *pgxpool.Pool `name:"rw"`
			},
			) {
				in.DB.Close()
			}),
		)),

		// setup read-only *sql.DB stdlib connection pool
		fx.Provide(fx.Annotate(New,
			fx.ParamTags(`name:"ro"`, `optional:"true"`, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"ro"`),
			fx.OnStop(func(ctx context.Context, in struct {
				fx.In
				DB *sql.DB `name:"ro"`
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
			fx.ParamTags(`name:"rw"`, `optional:"true"`, `optional:"true"`, `optional:"true"`),
			fx.ResultTags(`name:"rw"`),
			fx.OnStop(func(ctx context.Context, in struct {
				fx.In
				DB *sql.DB `name:"rw"`
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
		fx.Provide(fx.Annotate(func(rw *sql.DB) *sql.DB { return rw }, fx.ParamTags(`name:"rw"`))),
		fx.Provide(fx.Annotate(func(rw *pgxpool.Pool) *pgxpool.Pool { return rw }, fx.ParamTags(`name:"rw"`))),
		fx.Provide(fx.Annotate(func(rw *pgxpool.Config) *pgxpool.Config { return rw }, fx.ParamTags(`name:"rw"`))),

		// provide test with the ability to request a read-write tx that automatically rolls back and discourages
		// sequential scans for spotting expensive (non indexe) queries more easily.
		fx.Provide(func(lcl fx.Lifecycle, rw *pgxpool.Pool) (pgx.Tx, error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tx, err := rw.Begin(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to begin tx: %w", err)
			}
			// make it more likely to not do scans, so we can test index use more easily.
			if _, err := tx.Exec(ctx, `SET SESSION enable_seqscan = off`); err != nil {
				return nil, fmt.Errorf("failed to set: enable_seqscan = off: %w", err)
			}

			// rollback the tx after each test, we don't consider ErrTxClosed an error so test
			// CAN commit and assert the result whitout erroring at the end of the test.
			lcl.Append(fx.StopHook(func(ctx context.Context) error {
				if err := tx.Rollback(ctx); !errors.Is(err, pgx.ErrTxClosed) {
					return err //nolint: wrapcheck
				}

				return nil
			}))

			return tx, nil
		}),
	)
}
