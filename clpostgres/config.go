package clpostgres

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

// Config configures the code in this package.
type Config struct {
	// DatabaseName names the database the connection will be made to
	DatabaseName string `env:"DATABASE_NAME" envDefault:"postgres"`
	// ReadWriteHostname endpoint allows configuration of a endpoint that can read and write
	ReadWriteHostname string `env:"RW_HOSTNAME" envDefault:"localhost"`
	// ReadOnlyHostname endpoint allows configuration of a endpoint that can read and write
	ReadOnlyHostname string `env:"RW_HOSTNAME" envDefault:"localhost"`
	// Port for to the database connection(s)
	Port int `env:"PORT" envDefault:"5432"`
	// Username configures the username to connect to the postgres instance
	Username string `env:"USERNAME" envDefault:"postgres"`
	// Password configures the postgres password for authenticating with the instance
	Password string `env:"PASSWORD"`
	// IamAuthTimeout bounds the time it takes to geht the IAM auth token
	IamAuthTimeout time.Duration `env:"IAM_AUTH_TIMEOUT" envDefault:"100ms"`
	// PgxLogLevel is provided to pgx to determine the level of logging of postgres interactions
	PgxLogLevel string `env:"PGX_LOG_LEVEL" envDefault:"info"`
	// SchemaName sets the schema to which the connections search_path will be set
	SchemaName string `env:"SCHEMA_NAME" envDefault:"public"`
	// TemporarySchemaName can be set to non-empty to cause schema name to indicate it is temporary schema
	TemporarySchemaName string `env:"TEMPORARY_SCHEMA_NAME"`
	// SSLMode sets tls encryption on the database connection
	SSLMode string `env:"SSL_MODE" envDefault:"disable"`
}

// NewReadOnlyConfig constructs a config for a read-only database connecion. The aws config is optional
// and is only used when IamAuth option is set.
func NewReadOnlyConfig(cfg Config, logs *zap.Logger) (*pgxpool.Config, error) {
	return newPoolConfig(cfg, logs.Named("ro"), cfg.ReadOnlyHostname)
}

// NewReadWriteConfig constructs a config for a read-write database connecion. The aws config is optional
// and only used when the IamAuth option is set
func NewReadWriteConfig(cfg Config, logs *zap.Logger) (*pgxpool.Config, error) {
	return newPoolConfig(cfg, logs.Named("rw"), cfg.ReadWriteHostname)
}

// newPoolConfig will turn environment configuration in a way that allows
// database credentials to be provided
func newPoolConfig(cfg Config, logs *zap.Logger, ep string) (pcfg *pgxpool.Config, err error) {
	connString := fmt.Sprintf(`postgres://%s:%s@%s:%d/%s?application_name=cl.%s&sslmode=%s`,
		cfg.Username, url.QueryEscape(cfg.Password), ep, cfg.Port, cfg.DatabaseName, cfg.DatabaseName, cfg.SSLMode)
	pcfg, err = pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from conn string: %w", err)
	}

	ll, err := tracelog.LogLevelFromString(cfg.PgxLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to determine pgx log level from '%s': %w", cfg.PgxLogLevel, err)
	}

	// we use a tracer to log all interactions with the database
	pcfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   NewLogger(logs),
		LogLevel: ll}

	// if a temporary schema prefix is configured we assume to customize the connections search
	// path wich will trigger the migration code to setup in a temporary schema.
	if cfg.TemporarySchemaName != "" {
		pcfg.ConnConfig.RuntimeParams["search_path"] = cfg.TemporarySchemaName
	}

	return
}
