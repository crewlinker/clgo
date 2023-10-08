// Package clpgmigrate implements the postgres migration logic.
package clpgmigrate

// Config configures the packaged logic.
type Config struct {
	// TemporaryDatabase can be set to cause the logic to create a random database name and initialize
	// it when running auto-migration. This is mostly useful for automated tests
	TemporaryDatabase bool `env:"TEMPORARY_DATABASE" envDefault:"false"`
	// AutoMigration can be set to true to cause the logic to automatically run migrations when started. This
	// is mostly useful for automated tests.
	AutoMigration bool `env:"AUTO_MIGRATION" envDefault:"false"`
	// the sql being generated for creating the temporary database
	CreateDatabaseFormat string `env:"CREATE_DATABASE_FORMAT" envDefault:"CREATE DATABASE %s"`
	// the sql being generated for dropping the temporary database
	DropDatabaseFormat string `env:"DROP_DATABASE_FORMAT" envDefault:"DROP DATABASE %s (force)"`
}

// modulename for consistent config parsing.
const moduleName = "clpgmigrate"
