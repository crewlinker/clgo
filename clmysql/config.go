package clmysql

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// NewReadOnlyConfig constructs a config for a read-only database connecion. The aws config is optional
// and is only used when IamAuth option is set.
func NewReadOnlyConfig(cfg Config, logs *zap.Logger) (*mysql.Config, error) {
	return newConfig(cfg, logs.Named("ro"), cfg.ReadOnlyDataSourceName)
}

// NewReadWriteConfig constructs a config for a read-write database connecion. The aws config is optional
// and only used when the IamAuth option is set.
func NewReadWriteConfig(cfg Config, logs *zap.Logger) (*mysql.Config, error) {
	return newConfig(cfg, logs.Named("rw"), cfg.ReadWriteDataSourceName)
}

// newConfig will turn environment configuration in a way that allows
// database credentials to be provided.
func newConfig(_ Config, logs *zap.Logger, dsName string) (*mysql.Config, error) {
	mycfg, err := mysql.ParseDSN(dsName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data source name: %w", err)
	}

	logs.Info("initialized postgres connection config",
		zap.Any("runtime_params", mycfg.Params),
		zap.String("tls_config", mycfg.TLSConfig),
		zap.String("user", mycfg.User),
		zap.String("db_name", mycfg.DBName),
		zap.String("addr", mycfg.Addr),
		zap.String("net", mycfg.Net))

	return mycfg, nil
}
