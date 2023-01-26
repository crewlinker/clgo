package clpostgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// New inits a stdlib sql connection. Any other dependency can optionall ybe provided as migrated
// to force it's lifecycle to be run before the database is connected. This is mostly usefull to
// run migration logic (such as initializing the database)
func New(pcfg *pgxpool.Config, m *Migrater) (db *sql.DB) {
	c := stdlib.GetConnector(*pcfg.ConnConfig)
	db = sql.OpenDB(c)
	return
}

// moduleName for naming conventions
const moduleName = "clpostgres"

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

	// provide read/write configuration
	fx.Provide(fx.Annotate(NewReadOnlyConfig,
		fx.ParamTags(``, ``, `optional:"true"`), fx.ResultTags(`name:"ro"`))),
	fx.Provide(fx.Annotate(NewReadWriteConfig,
		fx.ParamTags(``, ``, `optional:"true"`), fx.ResultTags(`name:"rw"`))),

	// setup read-only connection
	fx.Provide(fx.Annotate(New,
		fx.ParamTags(`name:"ro"`, `optional:"true"`), fx.ResultTags(`name:"ro"`),
		fx.OnStart(func(ctx context.Context, in struct {
			fx.In
			DB *sql.DB `name:"ro"`
		}) error {
			return in.DB.PingContext(ctx)
		}),
		fx.OnStop(func(ctx context.Context, in struct {
			fx.In
			DB *sql.DB `name:"ro"`
		}) error {
			return in.DB.Close()
		}),
	)),

	// setup read-write connection
	fx.Provide(fx.Annotate(New,
		fx.ParamTags(`name:"rw"`, `optional:"true"`), fx.ResultTags(`name:"rw"`),
		fx.OnStart(func(ctx context.Context, in struct {
			fx.In
			DB *sql.DB `name:"rw"`
		}) error {
			return in.DB.PingContext(ctx)
		}),
		fx.OnStop(func(ctx context.Context, in struct {
			fx.In
			DB *sql.DB `name:"rw"`
		}) error {
			return in.DB.Close()
		}),
	)),
)

// Test configures the DI for a test environment
var Test = fx.Options(Prod,
	// Provide migrater which will now always run before connecting. If auto-migrate and temporary database
	// configuration is set this can provide a fully isolated and migrated database for each test.
	fx.Provide(fx.Annotate(
		NewMigrater,
		fx.OnStart(func(ctx context.Context, m *Migrater) error { return m.Migrate(ctx) }),
		fx.OnStop(func(ctx context.Context, m *Migrater) error { return m.DropDatabase(ctx) }),
		fx.ParamTags(``, ``, `name:"rw"`)),
	),

	// we re-provide the read-write sql db as an unnamed *sql.DB and config because that is
	// what we usually want in tests.
	fx.Provide(fx.Annotate(func(rw *sql.DB) *sql.DB { return rw }, fx.ParamTags(`name:"rw"`))),
	fx.Provide(fx.Annotate(func(rw *pgxpool.Config) *pgxpool.Config { return rw }, fx.ParamTags(`name:"rw"`))),
)
