package clpgmigrate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/sqltool"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// alwaysValidateMigrateDir can be provided as a migration dir to disable checksum validation. This
// is useful for unit tests that don't need this feature for quick iteration.
type alwaysValidateMigrateDir struct{ *sqltool.GolangMigrateDir }

// Checksum implements the logic for re-calculate the directories checksum. But instead we return
// the checksum in the checksum file to make the validation logic always pass.
func (dir alwaysValidateMigrateDir) Checksum() (migrate.HashFile, error) {
	file, err := dir.Open(migrate.HashFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open hash file: %w", err)
	}

	defer file.Close()

	byt, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read all from hash file: %w", err)
	}

	var fh migrate.HashFile
	if err := fh.UnmarshalText(byt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hash file: %w", err)
	}

	return fh, nil
}

// VersionMigrated configures the di for testing with a temporary database and auto-migration of a directory.
func VersionMigrated(migrationDir string, disableValidation bool) fx.Option {
	return fx.Options(
		// provide the environment configuration
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),

		// Provide migrater which will now always run before connecting using versioned steps
		fx.Provide(fx.Annotate(
			NewVersionMigrater,
			fx.As(new(clpostgres.Migrater)),
			fx.OnStart(func(ctx context.Context, m clpostgres.Migrater) error { return m.Migrate(ctx) }), //nolint:wrapcheck
			fx.OnStop(func(ctx context.Context, m clpostgres.Migrater) error { return m.Reset(ctx) }),    //nolint:wrapcheck
			fx.ParamTags(``, ``, `name:"rw"`, `name:"ro"`)),
		),

		// For tests we want temporary database and auto-migratino
		fx.Decorate(func(c Config) Config {
			c.TemporaryDatabase = true
			c.AutoMigration = true

			return c
		}),
		// provide the optional configuration for a migration dir
		fx.Provide(func() (migrate.Dir, error) {
			dir, err := sqltool.NewGolangMigrateDir(migrationDir)
			if err != nil {
				return nil, fmt.Errorf("failed to init golang migrate dir: %w", err)
			}

			if disableValidation {
				return alwaysValidateMigrateDir{GolangMigrateDir: dir}, nil
			}

			return dir, nil
		}),
	)
}

// VersionMigrater allows programmatic migration of a database schema using versioned sql steps. Mostly used
// in testing and local development to provide fully isolated databases.
type VersionMigrater struct {
	baseMigrater

	dir migrate.Dir
}

// NewVersionMigrater inits the migrater.
func NewVersionMigrater(
	cfg Config,
	logs *zap.Logger,
	rwcfg *pgxpool.Config,
	rocfg *pgxpool.Config,
	dir migrate.Dir,
) (*VersionMigrater, error) {
	mig := &VersionMigrater{
		baseMigrater: baseMigrater{
			cfg:   cfg,
			dbcfg: rwcfg,
			logs:  logs.Named("migrater"),
		}, dir: dir,
	}

	return mig, mig.baseMigrater.init(rocfg)
}

// Migrate initializes the schema.
func (m VersionMigrater) Migrate(ctx context.Context) error {
	if err := m.baseMigrater.setup(ctx); err != nil {
		return err
	}

	if !m.cfg.AutoMigration {
		m.logs.Info("auto-migration disabled, expect database to be provisioned already")

		return nil // without auto-migration enabled, there is nothing left to do
	}

	checksum, err := m.dir.Checksum()
	if err != nil {
		return fmt.Errorf("failed to determine local dir checksum: %w", err)
	}

	m.logs.Info("auto-migrating from migrate dir",
		zap.String("checksum", checksum.Sum()))

	sqldb := stdlib.OpenDB(*m.dbcfg.ConnConfig)
	defer sqldb.Close()

	if err := sqldb.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping connection: %w", err)
	}

	drv, err := postgres.Open(sqldb)
	if err != nil {
		return fmt.Errorf("failed to init atlas driver: %w", err)
	}

	exec, err := migrate.NewExecutor(drv, m.dir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return fmt.Errorf("failed to init executor: %w", err)
	}

	if err := exec.ExecuteN(ctx, -1); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}

// Reset drops the migrated state.
func (m VersionMigrater) Reset(ctx context.Context) error {
	return m.baseMigrater.reset(ctx)
}
