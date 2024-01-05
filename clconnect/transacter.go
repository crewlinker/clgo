package clconnect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clpostgres/cltx"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ROTransacter provides a database transaction in the context.
type ROTransacter struct {
	cfg  Config
	logs *zap.Logger
	ro   *pgxpool.Pool
	connect.Interceptor
}

// NewROTransacter inits the Transacter.
func NewROTransacter(cfg Config, logs *zap.Logger, ro *pgxpool.Pool) *ROTransacter {
	intr := &ROTransacter{cfg: cfg, logs: logs.Named("ro_transacter"), ro: ro}
	intr.Interceptor = connect.UnaryInterceptorFunc(intr.intercept)

	return intr
}

func (l ROTransacter) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return txIntercept(ctx, l.logs, req, l.ro, next, pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		})
	})
}

// RWTransacter provides a database transaction in the context.
type RWTransacter struct {
	cfg  Config
	logs *zap.Logger
	rw   *pgxpool.Pool
	connect.Interceptor
}

// NewRWTransacter inits the Transacter.
func NewRWTransacter(cfg Config, logs *zap.Logger, rw *pgxpool.Pool) *RWTransacter {
	intr := &RWTransacter{cfg: cfg, logs: logs.Named("rw_transacter"), rw: rw}
	intr.Interceptor = connect.UnaryInterceptorFunc(intr.intercept)

	return intr
}

func (l RWTransacter) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return txIntercept(ctx, l.logs, req, l.rw, next, pgx.TxOptions{})
	})
}

func txIntercept(
	ctx context.Context,
	logs *zap.Logger,
	req connect.AnyRequest,
	db *pgxpool.Pool,
	next connect.UnaryFunc,
	opts pgx.TxOptions,
) (connect.AnyResponse, error) {
	logs = clzap.Log(ctx, logs)

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if rberr := tx.Rollback(ctx); rberr != nil {
			logs.Error("failed to rollback tx", zap.Error(err))
		}
	}()

	ctx = cltx.WithTx(ctx, tx)

	resp, err := next(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil &&
		!errors.Is(err, pgx.ErrTxCommitRollback) { // is rolled back, that is fine
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return resp, nil
}
