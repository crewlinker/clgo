package clpgmigrate

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// GooseMigrateDir types a filesytem that holds the full path to the goose migrations.
type GooseMigrateDir struct{ fs.FS }

// GooseMigrater uses a goose strategy to migrate.
type GooseMigrater struct {
	baseMigrater
	dir fs.FS
}

// NewGooseMigrater inits the migrater.
func NewGooseMigrater(
	cfg Config,
	logs *zap.Logger,
	rwcfg *pgxpool.Config,
	rocfg *pgxpool.Config,
	dir GooseMigrateDir,
) (*GooseMigrater, error) {
	mig := &GooseMigrater{
		baseMigrater: baseMigrater{
			cfg:   cfg,
			dbcfg: rwcfg,
			logs:  logs.Named("goose_migrater"),
		},
		dir: fs.FS(dir),
	}

	goose.SetLogger(zap.NewStdLog(mig.logs))

	return mig, mig.baseMigrater.init(rocfg)
}

// Migrate initializes the schema.
func (m GooseMigrater) Migrate(ctx context.Context) error {
	if err := m.baseMigrater.setup(ctx); err != nil {
		return err
	}

	sqldb := stdlib.OpenDB(*m.dbcfg.ConnConfig)
	defer sqldb.Close()

	if err := sqldb.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping connection: %w", err)
	}

	prov, err := goose.NewProvider(goose.DialectPostgres, sqldb, m.dir)
	if err != nil {
		return fmt.Errorf("failed to init goose provider: %w", err)
	}

	if _, err := prov.Up(ctx); err != nil {
		return fmt.Errorf("failed to run goose: %w", err)
	}

	return nil
}

// Reset drops the schema.
func (m GooseMigrater) Reset(ctx context.Context) error {
	return m.baseMigrater.reset(ctx)
}

// GooseMigrated configures the di for using snapshot migrations.
func GooseMigrated(migrateDir fs.FS, cnff ...func(*Config)) fx.Option {
	return fx.Options(
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),

		fx.Supply(GooseMigrateDir{FS: migrateDir}),

		// Provide migrater which will now always run before connecting using versioned steps
		fx.Provide(fx.Annotate(
			NewGooseMigrater,
			fx.As(new(clpostgres.Migrater)),
			fx.OnStart(func(ctx context.Context, m clpostgres.Migrater) error { return m.Migrate(ctx) }),
			fx.OnStop(func(ctx context.Context, m clpostgres.Migrater) error { return m.Reset(ctx) }),
			fx.ParamTags(``, ``, `name:"rw"`, `name:"ro"`)),
		),

		// For tests we want temporary database and auto-migratino
		fx.Decorate(func(c Config) Config {
			c.TemporaryDatabase = true
			c.AutoMigration = true

			// allow config to be overwritten
			if len(cnff) > 0 {
				cnff[0](&c)
			}

			return c
		}),
	)
}
