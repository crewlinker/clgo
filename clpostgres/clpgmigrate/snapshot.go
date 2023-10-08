package clpgmigrate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/jackc/pgx/v5/pgxpool"
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
	rwcfg *pgxpool.Config,
	rocfg *pgxpool.Config,
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

	tmpf, err := os.CreateTemp("", "snapshot_migrater_*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	defer tmpf.Close()

	if _, err := io.Copy(tmpf, bytes.NewBuffer(m.snapshot)); err != nil {
		return fmt.Errorf("failed to write snapshot to temp file: %w", err)
	}

	exe, err := exec.LookPath("psql")
	if err != nil {
		return fmt.Errorf("failed to lookup 'psql': %w", err)
	}

	errb := bytes.NewBuffer(nil)

	cmd := exec.CommandContext(ctx,
		exe,
		"-h", m.dbcfg.ConnConfig.Host,
		"-p", fmt.Sprint(m.dbcfg.ConnConfig.Port),
		"-U", m.dbcfg.ConnConfig.User,
		"-d", m.dbcfg.ConnConfig.Database,
		"-a", "-f", tmpf.Name())
	cmd.Env = []string{"PGPASSWORD=" + m.dbcfg.ConnConfig.Password}
	cmd.Stderr = errb

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute: %w: %s", err, errb.String())
	}

	if err := os.Remove(tmpf.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
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
	)
}
