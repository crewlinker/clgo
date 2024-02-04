package clmymigrate

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// baseMigrater implements shared logic between the migraters.
type baseMigrater struct {
	cfg       Config
	dbcfg     *mysql.Config
	logs      *zap.Logger
	databases struct {
		original *mysql.Config
		temp     string
	}
}

// init performs initialization of the migrater.
func (m *baseMigrater) init(rocfg *mysql.Config) error {
	if m.cfg.TemporaryDatabase {
		var rngd [6]byte
		if _, err := rand.Read(rngd[:]); err != nil {
			return fmt.Errorf("failed to read random bytes for temp name: %w", err)
		}

		// if a temporary database is requested, we replace the database name in the connection string
		// but keep the original for a bootstrap connection later.
		m.databases.original = m.dbcfg.Clone()

		// m.databases.original = m.dbcfg.Copy()
		m.databases.temp = fmt.Sprintf("temp_%x_%s", rngd, m.databases.original.DBName)
		m.dbcfg.DBName = m.databases.temp
		// also change in read-only connection config, or the read-only is reading from different database
		rocfg.DBName = m.databases.temp
	}

	return nil
}

// bootstrapRun will init a dedicated bootstrap connection.
func (m baseMigrater) bootstrapRun(ctx context.Context, runf func(context.Context, *sql.DB) error) error {
	return func(ctx context.Context) error {
		connr, err := mysql.NewConnector(m.databases.original)
		if err != nil {
			return fmt.Errorf("failed to init connector: %w", err)
		}

		db := sql.OpenDB(connr)
		defer db.Close()

		if err := runf(ctx, db); err != nil {
			return fmt.Errorf("failed to run: %w", err)
		}

		return nil
	}(ctx)
}

// setup shared migrater logic.
func (m baseMigrater) setup(ctx context.Context) error {
	if m.cfg.TemporaryDatabase {
		m.logs.Info("enabled temporary database option, creating database",
			zap.String("bootstrap_database_name", m.databases.original.DBName),
			zap.String("database_name", m.databases.temp))

		// with a temporary database we create it using a bootstrap connection
		if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
			if _, err := db.ExecContext(ctx, fmt.Sprintf(m.cfg.CreateDatabaseFormat, m.databases.temp)); err != nil {
				return fmt.Errorf("failed to exec create database sql: %w", err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to run from bootstrap database: %w", err)
		}
	}

	return nil
}

// reset the database.
func (m baseMigrater) reset(ctx context.Context) error {
	if m.databases.temp == "" {
		return nil // not temporary database, so nothing to do on shutdown
	}

	m.logs.Info("temporary database enabled, dropping database",
		zap.String("bootstrap_database_name", m.databases.original.DBName),
		zap.String("database_name", m.databases.temp))

	if err := m.bootstrapRun(ctx, func(ctx context.Context, db *sql.DB) error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(m.cfg.DropDatabaseFormat, m.databases.temp))
		if err != nil {
			return fmt.Errorf("failed to execute database drop: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed turn from bootstrap database: %w", err)
	}

	return nil
}
