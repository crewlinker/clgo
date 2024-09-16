package clpgmigrate

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// TemplateDatabaseName holds the name of a template database.
type TemplateDatabaseName string

// TemplateMigrater uses a template database to migrate.
type TemplateMigrater struct {
	baseMigrater
	templateName string
}

// small utility to run some sql and disconnect.
func runSQL(ctx context.Context, connURL *url.URL, stmts ...string) error {
	conn, err := pgx.Connect(ctx, connURL.String())
	if err != nil {
		return fmt.Errorf("failed init bootstrap connection: %w", err)
	}

	for _, stmt := range stmts {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to run sql (%s): %w", stmt, err)
		}
	}
	if err = conn.Close(ctx); err != nil {
		return fmt.Errorf("failed to close bootstrap connection: %w", err)
	}

	return nil
}

// SetupTemplateDatabaseFromSnapshot can be called early to setup a template database.
func SetupTemplateDatabaseFromSnapshot(
	ctx context.Context,
	bootstrapConnString,
	templateDatabaseName,
	snapshotPath string,
) error {
	connStringURL, err := url.Parse(bootstrapConnString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	psqlExe, err := exec.LookPath("psql")
	if err != nil {
		return fmt.Errorf("failed to lookup 'psql': %w", err)
	}

	// create the template database
	if err := runSQL(ctx, connStringURL,
		`CREATE DATABASE `+templateDatabaseName+` IS_TEMPLATE true`); err != nil {
		return err
	}

	templConnString := *connStringURL
	templConnString.Path = templateDatabaseName

	// populate from snapshot
	errb := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx,
		psqlExe,
		"-d", templConnString.String(),
		"-a", "--set", "ON_ERROR_STOP=on",
		"-f", snapshotPath)
	cmd.Stderr = errb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run psql to apply snapshot to template database: %w: %s", err, errb.String())
	}

	// mark as not allowing connections
	if err := runSQL(ctx, connStringURL,
		`ALTER DATABASE `+templateDatabaseName+` ALLOW_CONNECTIONS false`); err != nil {
		return err
	}

	return nil
}

// TeardownTemplateDatabase can called late in the test suite to drop the template database.
func TeardownTemplateDatabase(
	ctx context.Context,
	bootstrapConnString,
	templateDatabaseName string,
) error {
	connStringURL, err := url.Parse(bootstrapConnString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// drop the template database.
	if err := runSQL(ctx, connStringURL,
		`ALTER DATABASE `+templateDatabaseName+` IS_TEMPLATE false`,
		`DROP DATABASE IF EXISTS `+templateDatabaseName); err != nil {
		return err
	}

	return nil
}

// NewTemplateMigrater inits the migrater.
func NewTemplateMigrater(
	cfg Config,
	logs *zap.Logger,
	rwcfg *pgxpool.Config,
	tmpl TemplateDatabaseName,
) (mig *TemplateMigrater, err error) {
	mig = &TemplateMigrater{
		baseMigrater: baseMigrater{
			cfg:   cfg,
			dbcfg: rwcfg,
			logs:  logs.Named("migrater"),
		},
		templateName: string(tmpl),
	}

	// Important bit 1: this migrater expects a template database to exist that we clone.
	mig.baseMigrater.cfg.CreateDatabaseFormat += " TEMPLATE " + mig.templateName

	if err := mig.baseMigrater.init(rwcfg); err != nil {
		return nil, fmt.Errorf("failed to init base migrater: %w", err)
	}

	// Important bit 2: this migrater runs migrations in the Fx init phase, so tests can more easily create a tx.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.InitMigrateTimeout)
	defer cancel()

	if err := mig.baseMigrater.setup(ctx); err != nil {
		return nil, fmt.Errorf("failed to perform base migrater: %w", err)
	}

	return mig, nil
}

// Migrate initializes the schema. This migrater does the migration in the init phase.
func (m TemplateMigrater) Migrate(ctx context.Context) error {
	return nil
}

// Reset drops the schema.
func (m TemplateMigrater) Reset(ctx context.Context) error {
	if err := m.baseMigrater.reset(ctx); err != nil {
		return fmt.Errorf("failed to reset mase migrater: %w", err)
	}

	return nil
}

// TemplateMigrated configures the di for using template based migrations.
func TemplateMigrated() fx.Option {
	return fx.Options(
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),

		// Provide migrater which will now always run before connecting using versioned steps
		fx.Provide(fx.Annotate(
			NewTemplateMigrater,
			fx.As(new(clpostgres.Migrater)),
			// NOTE: unlike other migraters we don't run the migrate logic in the lifecycle phase.
			fx.OnStop(func(ctx context.Context, m clpostgres.Migrater) error { return m.Reset(ctx) }),
			fx.ParamTags(``, ``, `name:"rw"`)),
		),

		// For tests we want temporary database and auto-migratino
		fx.Decorate(func(c Config) Config {
			c.TemporaryDatabase = true
			c.AutoMigration = true

			return c
		}),
	)
}
