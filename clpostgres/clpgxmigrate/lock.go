package clpgxmigrate

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Locker interface {
	Lock(ctx context.Context, conn *pgx.Conn) error
	Unlock(ctx context.Context, conn *pgx.Conn) error
}

func NewPostgresAdvisoryLocker(num int64) Locker {
	return &pgAdvisoryLocker{num: num}
}

const advisoryLockNumber = int64(8808257919277131071)

type pgAdvisoryLocker struct{ num int64 }

func (l *pgAdvisoryLocker) Lock(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "select pg_advisory_lock($1)", l.num); err != nil {
		return fmt.Errorf("failed to select pg_advisory_lock: %w", err)
	}

	return nil
}

func (l *pgAdvisoryLocker) Unlock(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "select pg_advisory_unlock($1)", l.num); err != nil {
		return fmt.Errorf("failed to select pg_advisory_unlock: %w", err)
	}

	return nil
}
