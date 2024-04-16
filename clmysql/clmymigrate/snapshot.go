package clmymigrate

import (
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clmysql"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// SnapshotPath types a string to hold the full snapshot sql.
type SnapshotPath string

// SnapshotMigrater uses a snapshot strategy to migrate.
type SnapshotMigrater struct {
	baseMigrater
	snapshot []byte
}

// NewSnaphotMigrater inits the migrater.
func NewSnaphotMigrater(
	cfg Config,
	logs *zap.Logger,
	rwcfg *mysql.Config,
	rocfg *mysql.Config,
	snapp SnapshotPath,
) (*SnapshotMigrater, error) {
	var err error

	mig := &SnapshotMigrater{
		baseMigrater: baseMigrater{
			cfg:   cfg,
			dbcfg: rwcfg,
			logs:  logs.Named("migrater"),
		},
	}

	mig.snapshot, err = os.ReadFile(string(snapp))
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot: %w", err)
	}

	return mig, mig.baseMigrater.init(rocfg)
}

// Migrate initializes the schema.
func (m SnapshotMigrater) Migrate(ctx context.Context) error {
	if err := m.baseMigrater.setup(ctx); err != nil {
		return err
	}

	exe, err := exec.LookPath("mysql")
	if err != nil {
		return fmt.Errorf("failed to lookup 'psql': %w", err)
	}

	addr, err := netip.ParseAddrPort(m.dbcfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to parse addr/port: %w", err)
	}

	errb := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, exe,
		"-h", addr.Addr().String(),
		"-P", strconv.FormatUint(uint64(addr.Port()), 10),
		"-u", m.dbcfg.User,
		"-p"+m.dbcfg.Passwd, // no space
		m.dbcfg.DBName)
	cmd.Stdin = bytes.NewBuffer(m.snapshot)
	cmd.Stderr = errb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute: %w: %s", err, errb.String())
	}

	return nil
}

// Reset drops the schema.
func (m SnapshotMigrater) Reset(ctx context.Context) error {
	return m.baseMigrater.reset(ctx)
}

// SnapshotMigrated configures the di for using snapshot migrations.
func SnapshotMigrated(sqlFile string) fx.Option {
	return fx.Options(
		clconfig.Provide[Config](strings.ToUpper(moduleName)+"_"),

		fx.Supply(SnapshotPath(sqlFile)),

		// Provide migrater which will now always run before connecting using versioned steps
		fx.Provide(fx.Annotate(
			NewSnaphotMigrater,
			fx.As(new(clmysql.Migrater)),
			fx.OnStart(func(ctx context.Context, m clmysql.Migrater) error { return m.Migrate(ctx) }),
			fx.OnStop(func(ctx context.Context, m clmysql.Migrater) error { return m.Reset(ctx) }),
			fx.ParamTags(``, ``, `name:"my_rw"`, `name:"my_ro"`)),
		),

		// For tests we want temporary database and auto-migratino
		fx.Decorate(func(c Config) Config {
			c.TemporaryDatabase = true
			c.AutoMigration = true

			return c
		}),
	)
}
