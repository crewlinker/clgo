package clpgxmigrate

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
)

type Provider struct {
	conn *pgx.Conn
	*opts
}

func New(conn *pgx.Conn, os ...Option) *Provider {
	return NewProvider(conn, os...)
}

func NewProvider(
	conn *pgx.Conn,
	os ...Option,
) *Provider {
	return &Provider{conn, applyOpts(os...)}
}

func (p *Provider) UpdateSchemaVersion(ctx context.Context, version int64) error {
	if _, err := p.conn.Exec(ctx,
		fmt.Sprintf(`UPDATE "%s"."%s" SET version=$1`, p.versionSchemaName, p.versionTableName),
		version); err != nil {
		return fmt.Errorf("failed to update version table with current version: %w", err)
	}

	return nil
}

func (p *Provider) ReadSchemaVersion(ctx context.Context) (version int64, err error) {
	if err := p.conn.QueryRow(ctx,
		fmt.Sprintf(`SELECT version FROM "%s"."%s"`, p.versionSchemaName, p.versionTableName),
	).Scan(&version); err != nil {
		return -1, fmt.Errorf("failed to query and scan current version from version table: %w", err)
	}

	return version, nil
}

func (p *Provider) initialize(ctx context.Context) (done func(error) error, err error) {
	if err := p.locker.Lock(ctx, p.conn); err != nil {
		return nil, fmt.Errorf("failed to lock: %w", err)
	}

	if _, err := p.conn.Exec(ctx,
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`,
			p.versionSchemaName),
	); err != nil {
		return nil, fmt.Errorf("failed to ensure the schema version table exists: %w", err)
	}

	if _, err := p.conn.Exec(ctx,
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (version BIGINT NOT NULL)`,
			p.versionSchemaName, p.versionTableName),
	); err != nil {
		return nil, fmt.Errorf("failed to ensure the schema version table exists: %w", err)
	}

	if _, err := p.conn.Exec(ctx, fmt.Sprintf(`
		INSERT INTO "%s"."%s" (version)
		SELECT 0
		WHERE NOT EXISTS (
    		SELECT 1 FROM "%s"."%s"
		);
	`, p.versionSchemaName, p.versionTableName, p.versionSchemaName, p.versionTableName)); err != nil {
		return nil, fmt.Errorf("failed to ensure there is always (at least) a single version row: %w", err)
	}

	return func(berr error) error {
		if err := p.locker.Unlock(ctx, p.conn); err != nil {
			return errors.Join(berr, err)
		}

		return errors.Join(berr)
	}, nil
}

func VersionsToApply(versions []int64, current, target int64) (result []int64) {
	slices.Sort(versions)

	for _, version := range versions {
		if version <= current {
			continue
		}

		if version <= target {
			result = append(result, version)
		}
	}

	return
}

type Result struct {
	AppliedVersions []int64
}

func (p *Provider) Migrate(ctx context.Context, targetVersion int64) (*Result, error) {
	done, err := p.initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	currentVersion, err := p.ReadSchemaVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to determine current version: %w", err)
	}

	versions := p.coll.Versions()
	toApply := VersionsToApply(versions, currentVersion, targetVersion)
	if len(toApply) < 1 {
		p.logger.LogNoVersionToApply(ctx, versions, toApply, currentVersion, targetVersion)

		return &Result{}, nil
	}

	p.logger.LogVersionsToApply(ctx, versions, toApply, currentVersion, targetVersion)

	for _, version := range toApply {
		step, ok := p.coll.Step(version)
		if !ok {
			panic(fmt.Sprintf("clpgxmigrate: trying to apply a step that was never registered: %d", version))
		}

		p.logger.LogVersionApplyStart(ctx, version)

		if err := step.Apply(ctx, p.conn); err != nil {
			return nil, applyError(version, "failed to apply: %w", err)
		}

		p.logger.LogVersionApplyDone(ctx, version)

		if err := p.UpdateSchemaVersion(ctx, version); err != nil {
			return nil, fmt.Errorf("failed to update the schema version to: %d, it no longer reflects the actual version: %w",
				version, err)
		}
	}

	return &Result{AppliedVersions: toApply}, done(err)
}

type Status struct {
	CurrentVersion int64
}

func (p *Provider) Status(ctx context.Context) (status *Status, err error) {
	status = &Status{}

	if status.CurrentVersion, err = p.ReadSchemaVersion(ctx); err != nil {
		return nil, fmt.Errorf("failed to determine current version: %w", err)
	}

	return status, nil
}
