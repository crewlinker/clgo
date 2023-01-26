package clatlas

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"github.com/caarlos0/env/v6"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Config configures the atlas migration logic
type Config struct {
	// TemporaryDatabase can be set to true to enable the database to be created an migrated fully on every
	// run of every program. Mostly usefull for testing
	TemporaryDatabase bool `env:"TEMPORARY_DATABASE"`
	// MigrationsDir configures from what directory the migrations are read. By default it will read from
	// a directory called "migrations"
	MigrationsDir string `env:"MIGRATIONS_DIR" envDefault:"migrations"`
}

// Migrater allows programmatic migration of a database schema. Mostly used in testing with
// the TEMPORARY_DATABASE environment variable set. When this is the case it will replace the database name
// in the connection config with a randomly generated one and migrate it before any connecti on is made.
type Migrater struct {
	cfg   Config
	dbcfg *pgxpool.Config
	logs  *zap.Logger

	databases struct {
		original *pgxpool.Config
		temp     string
	}
}

// NewMigrator inits the migrater
func NewMigrater(cfg Config, logs *zap.Logger, dbcfg *pgxpool.Config) (*Migrater, error) {
	m := &Migrater{cfg: cfg, dbcfg: dbcfg, logs: logs.Named("migrater")}
	if !cfg.TemporaryDatabase {
		return m, nil
	}

	var rngd [6]byte
	if _, err := rand.Read(rngd[:]); err != nil {
		return nil, fmt.Errorf("failed to read random bytes for temp name: %w", err)
	}

	// if a temporary database is requested, we replace the database name in the connection string
	// but keep the original for a bootstrap connection later.
	m.databases.original = m.dbcfg.Copy()
	m.databases.temp = fmt.Sprintf("temp_%x_%s", rngd, m.databases.original.ConnConfig.Database)
	m.dbcfg.ConnConfig.Database = m.databases.temp
	return m, nil
}

// InitSchema initializes the schema
func (m Migrater) InitSchema(ctx context.Context) error {
	if m.databases.temp == "" {
		return nil // not temporary database, nothing to migrate
	}

	if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, m.databases.temp))
		return err
	}); err != nil {
		return fmt.Errorf("failed to bootstrap run: %w", err)
	}

	ldir, err := migrate.NewLocalDir(m.cfg.MigrationsDir)
	if err != nil {
		return fmt.Errorf("failed to init migration dir: %w", err)
	}

	checksum, err := ldir.Checksum()
	if err != nil {
		return fmt.Errorf("failed to determine local dir checksum: %w", err)
	}

	m.logs.Info("migrating from local directory",
		zap.String("dir", ldir.Path()),
		zap.String("checksum", checksum.Sum()))

	db := stdlib.OpenDB(*m.dbcfg.ConnConfig)
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping connection: %w", err)
	}

	drv, err := postgres.Open(db)
	if err != nil {
		return fmt.Errorf("failed to init atlas driver: %w", err)
	}

	exec, err := migrate.NewExecutor(drv, ldir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return fmt.Errorf("failed to init executor: %w", err)
	}

	if err := exec.ExecuteN(ctx, -1); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}

// DropSchema drops the schema
func (m Migrater) DropSchema(ctx context.Context) error {
	if m.databases.temp == "" {
		return nil // not temporary database, nothing to drop
	}

	if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE %s (force)`, m.databases.temp))
		return err
	}); err != nil {
		return fmt.Errorf("failed to bootstrap run: %w", err)
	}

	return nil
}

// bootstrapRun will init a dedicated bootstrap connection
func (m Migrater) bootstrapRun(ctx context.Context, runf func(context.Context, *sql.DB) error) error {
	return func(ctx context.Context) (err error) {
		db := stdlib.OpenDB(*m.databases.original.ConnConfig)
		defer db.Close()

		if err := runf(ctx, db); err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		return
	}(ctx)
}

// moduleName for naming conventions
const moduleName = "clatlas"

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
	// Provide migrater for code to allow migrations to be triggered
	fx.Provide(fx.Annotate(
		NewMigrater,
		fx.OnStart(func(ctx context.Context, m *Migrater) error { return m.InitSchema(ctx) }),
		fx.OnStop(func(ctx context.Context, m *Migrater) error { return m.DropSchema(ctx) }),
		fx.ParamTags(``, ``, `name:"rw"`)),
	),
)

// Test configures the DI for a test environment
var Test = fx.Options(Prod,

	// Re-export the migrater tagged as an interface to cause migrations to be run in tests
	fx.Provide(func(m *Migrater) clpostgres.Migrated { return m }),

	// specify a temporary schema in tests
	fx.Decorate(func(cfg Config) Config {
		cfg.TemporaryDatabase = true
		return cfg
	}),
)
