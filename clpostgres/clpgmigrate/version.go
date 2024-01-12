package clpgmigrate

import (
	"context"
	"fmt"
	"io"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// alwaysValidateMigrateDir can be provided as a migration dir to disable checksum validation. This
// is useful for unit tests that don't need this feature for quick iteration.
type alwaysValidateMigrateDir struct{ migrate.Dir }

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
func VersionMigrated(noChecksumValidate bool) fx.Option {
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
		// decorate dir to disable checksum
		fx.Decorate(func(dir migrate.Dir) (migrate.Dir, error) {
			if noChecksumValidate {
				return alwaysValidateMigrateDir{Dir: dir}, nil
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
			logs:  logs.Named("version_migrater"),
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

	exec, err := migrate.NewExecutor(drv, m.dir, migrate.NopRevisionReadWriter{},
		migrate.WithLogger(migrateLogger{logs: m.logs.Named("executer")}))
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

// migrateLogger logger.
type migrateLogger struct {
	logs *zap.Logger
}

func (l migrateLogger) Log(e migrate.LogEntry) {
	switch entry := e.(type) {
	case migrate.LogExecution:
		l.logs.Info("execution",
			zap.String("from", entry.From),
			zap.String("to", entry.To),
			zap.Strings("files", lo.Map(entry.Files, func(f migrate.File, _ int) string {
				return f.Name()
			})))
	case migrate.LogFile:
		l.logs.Info("file",
			zap.String("desc", entry.File.Desc()),
			zap.String("name", entry.File.Name()),
			zap.Int("skip", entry.Skip),
			zap.String("version", entry.Version),
		)
	case migrate.LogStmt:
		l.logs.Info("statement",
			zap.String("sql", entry.SQL),
		)
	case migrate.LogDone:
		l.logs.Info("done")
	case migrate.LogError:
		l.logs.Error("error",
			zap.Error(entry.Error),
			zap.String("sql", entry.SQL))
	}
}
