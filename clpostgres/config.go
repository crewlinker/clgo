package clpostgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/jackc/pgx/v5"
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
	ReadOnlyHostname string `env:"RO_HOSTNAME" envDefault:"localhost"`
	// Port for to the database connection(s)
	Port int `env:"PORT" envDefault:"5432"`
	// Username configures the username to connect to the postgres instance
	Username string `env:"USERNAME" envDefault:"postgres"`
	// Password configures the postgres password for authenticating with the instance
	Password string `env:"PASSWORD"`

	// ApplicationName allows the application to indicate its name so connections can be more easily debugged
	ApplicationName string `env:"APPLICATION_NAME" envDefault:"unknown"`
	// PgxLogLevel is provided to pgx to determine the level of logging of postgres interactions
	PgxLogLevel string `env:"PGX_LOG_LEVEL" envDefault:"info"`

	// SSLMode sets tls encryption on the database connection
	SSLMode string `env:"SSL_MODE" envDefault:"disable"`
	// IamAuthRegion will cause the password to be set to an IAM token for authentication
	IamAuthRegion string `env:"IAM_AUTH_REGION"`

	// PoolConnectionTimeout configures how long the pgx pool connect logic waits for the connection to establish
	PoolConnectionTimeout time.Duration `env:"POOL_CONNECTION_TIMEOUT" envDefault:"10s"`
}

// NewReadOnlyConfig constructs a config for a read-only database connecion. The aws config is optional
// and is only used when IamAuth option is set.
func NewReadOnlyConfig(cfg Config, logs *zap.Logger, awsc aws.Config) (*pgxpool.Config, error) {
	return newPoolConfig(cfg, logs.Named("ro"), cfg.ReadOnlyHostname, awsc)
}

// NewReadWriteConfig constructs a config for a read-write database connecion. The aws config is optional
// and only used when the IamAuth option is set.
func NewReadWriteConfig(cfg Config, logs *zap.Logger, awsc aws.Config) (*pgxpool.Config, error) {
	return newPoolConfig(cfg, logs.Named("rw"), cfg.ReadWriteHostname, awsc)
}

// error when invalid dep combo for config.
var errIAMAuthWithoutAWSConfig = errors.New("IAM auth requested but optional AWS config dependency not provided")

// newPoolConfig will turn environment configuration in a way that allows
// database credentials to be provided.
func newPoolConfig(cfg Config, logs *zap.Logger, host string, awsc aws.Config) (*pgxpool.Config, error) {
	connString := fmt.Sprintf(`postgres://%s:%s@%s/%s?application_name=%s&sslmode=%s`,
		cfg.Username, url.QueryEscape(cfg.Password), net.JoinHostPort(host, strconv.Itoa(cfg.Port)),
		cfg.DatabaseName, cfg.ApplicationName, cfg.SSLMode)

	pcfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from conn string: %w", err)
	}

	if cfg.IamAuthRegion != "" {
		if awsc.Credentials == nil {
			return nil, errIAMAuthWithoutAWSConfig
		}

		// For IAM Auth we need to build a token as a password on every connection attempt
		pcfg.BeforeConnect = func(ctx context.Context, pgc *pgx.ConnConfig) error {
			tok, err := buildIamAuthToken(ctx, logs, cfg.Port, cfg.Username, cfg.IamAuthRegion, awsc, host)
			if err != nil {
				return fmt.Errorf("failed to build iam token: %w", err)
			}

			pgc.Password = tok

			return nil
		}
	}

	lls, err := tracelog.LogLevelFromString(cfg.PgxLogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to determine pgx log level from '%s': %w", cfg.PgxLogLevel, err)
	}

	// we use a tracer to log all interactions with the database
	pcfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   NewLogger(pcfg, logs),
		LogLevel: lls,
	}

	logs.Info("initialized postgres connection config",
		zap.Any("runtime_params", pcfg.ConnConfig.RuntimeParams),
		zap.String("ssl_mode", cfg.SSLMode),
		zap.String("iam_auth_region", cfg.IamAuthRegion),
		zap.String("user", pcfg.ConnConfig.User),
		zap.String("database", pcfg.ConnConfig.Database),
		zap.String("host", pcfg.ConnConfig.Host),
		zap.Uint16("port", pcfg.ConnConfig.Port))

	return pcfg, nil
}

// buildIamAuthToken will construct a RDS proxy authentication token. We don't run this during the
// lifecycle phase so we timeout manually with our own context.
func buildIamAuthToken(
	ctx context.Context,
	logs *zap.Logger,
	port int,
	username, region string,
	awsc aws.Config,
	host string,
) (string, error) {
	ep := host + ":" + strconv.Itoa(port)

	logs.Info("building IAM auth token",
		zap.String("username", username),
		zap.String("region", region),
		zap.String("ep", ep))

	tok, err := auth.BuildAuthToken(ctx, ep, region, username, awsc.Credentials)
	if err != nil {
		return "", fmt.Errorf("underlying: %w", err)
	}

	return tok, nil
}
