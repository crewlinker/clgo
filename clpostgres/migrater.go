package clpostgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// Migrater allows programmatic migration of a database schema. Mostly used in testing and local development
// to provide fully isolated databases.
type Migrater struct {
	cfg   Config
	dbcfg *pgxpool.Config
	logs  *zap.Logger
	dir   migrate.Dir

	databases struct {
		original *pgxpool.Config
		temp     string
	}
}

// NewMigrater inits the migrater.
func NewMigrater(
	cfg Config,
	logs *zap.Logger,
	rwcfg *pgxpool.Config,
	rocfg *pgxpool.Config,
	dir migrate.Dir,
) (*Migrater, error) {
	mig := &Migrater{cfg: cfg, dbcfg: rwcfg, logs: logs.Named("migrater"), dir: dir}

	if cfg.TemporaryDatabase {
		var rngd [6]byte
		if _, err := rand.Read(rngd[:]); err != nil {
			return nil, fmt.Errorf("failed to read random bytes for temp name: %w", err)
		}

		// if a temporary database is requested, we replace the database name in the connection string
		// but keep the original for a bootstrap connection later.
		mig.databases.original = mig.dbcfg.Copy()
		mig.databases.temp = fmt.Sprintf("temp_%x_%s", rngd, mig.databases.original.ConnConfig.Database)
		mig.dbcfg.ConnConfig.Database = mig.databases.temp
		// also change in read-only connection config, or the read-only is reading from different database
		rocfg.ConnConfig.Database = mig.databases.temp
	}

	return mig, nil
}

// Migrate initializes the schema.
func (m Migrater) Migrate(ctx context.Context) error {
	if m.cfg.TemporaryDatabase {
		m.logs.Info("enabled temporary database option, creating database",
			zap.String("bootstrap_database_name", m.databases.original.ConnConfig.Database),
			zap.String("database_name", m.databases.temp))

		// with a temporary database we create it using a bootstrap connection
		if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s`, m.databases.temp)); err != nil {
				return fmt.Errorf("failed to exec context: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to bootstrap run: %w", err)
		}
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

// DropDatabase drops the schema.
func (m Migrater) DropDatabase(ctx context.Context) error {
	if m.databases.temp == "" {
		return nil // not temporary database, so nothing to do on shutdown
	}

	m.logs.Info("temporary database enabled, dropping database",
		zap.String("bootstrap_database_name", m.databases.original.ConnConfig.Database),
		zap.String("database_name", m.databases.temp))

	if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE %s (force)`, m.databases.temp))
		if err != nil {
			return fmt.Errorf("failed to exec: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to bootstrap run: %w", err)
	}

	return nil
}

// bootstrapRun will init a dedicated bootstrap connection.
func (m Migrater) bootstrapRun(ctx context.Context, runf func(context.Context, *sql.DB) error) error {
	return func(ctx context.Context) error {
		db := stdlib.OpenDB(*m.databases.original.ConnConfig)
		defer db.Close()

		if err := runf(ctx, db); err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		return nil
	}(ctx)
}
