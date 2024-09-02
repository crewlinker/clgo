package clpgxmigrate

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type StepFunc interface {
	~func(context.Context, pgx.Tx) error | ~func(context.Context, *pgx.Conn) error
}

func NewStep[T StepFunc](fn T) Step {
	switch f := any(fn).(type) {
	case func(context.Context, pgx.Tx) error:
		return step{txf: f}
	case func(context.Context, *pgx.Conn) error:
		return step{connf: f}
	default:
		panic(fmt.Sprintf("clpgxmigrate: invalid step type: %T", f))
	}
}

type Step interface {
	Apply(ctx context.Context, conn *pgx.Conn) error
}

type step struct {
	connf func(context.Context, *pgx.Conn) error
	txf   func(context.Context, pgx.Tx) error
}

func (s step) Apply(ctx context.Context, conn *pgx.Conn) error {
	switch {
	case s.connf != nil:
		return s.connf(ctx, conn)
	case s.txf != nil:
	default:
		panic("clpgxmigrate: invalid step, connection-based or a tx-based function must be provided")
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx) //nolint:errcheck

	if err := s.txf(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
