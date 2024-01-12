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
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// PgxROTransacter provides a database transaction in the context.
type PgxROTransacter struct {
	cfg  Config
	logs *zap.Logger
	ro   *pgxpool.Pool
	connect.Interceptor
}

// NewPgxROTransacter inits the Transacter.
func NewPgxROTransacter(cfg Config, logs *zap.Logger, ro *pgxpool.Pool) *PgxROTransacter {
	intr := &PgxROTransacter{cfg: cfg, logs: logs.Named("pgx_ro_transacter"), ro: ro}
	intr.Interceptor = connect.UnaryInterceptorFunc(intr.intercept)

	return intr
}

func (l PgxROTransacter) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return txPgxIntercept(ctx, l.logs, req, l.ro, next, pgx.TxOptions{
			AccessMode: pgx.ReadOnly,
		})
	})
}

// PgxRWTransacter provides a database transaction in the context.
type PgxRWTransacter struct {
	cfg  Config
	logs *zap.Logger
	rw   *pgxpool.Pool
	connect.Interceptor
}

// NewPgxRWTransacter inits the Transacter.
func NewPgxRWTransacter(cfg Config, logs *zap.Logger, rw *pgxpool.Pool) *PgxRWTransacter {
	intr := &PgxRWTransacter{cfg: cfg, logs: logs.Named("pgx_rw_transacter"), rw: rw}
	intr.Interceptor = connect.UnaryInterceptorFunc(intr.intercept)

	return intr
}

func (l PgxRWTransacter) intercept(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		return txPgxIntercept(ctx, l.logs, req, l.rw, next, pgx.TxOptions{})
	})
}

func txPgxIntercept(
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

	ctx = cltx.WithPgx(ctx, tx)

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

// ProvidePgxTransactors provides transactors for pgx transactions.
func ProvidePgxTransactors() fx.Option {
	return fx.Options(
		// database transactors
		fx.Provide(fx.Annotate(NewPgxROTransacter,
			fx.As(new(ROTransacter)),
			fx.ParamTags(``, ``, `name:"ro"`))),
		fx.Provide(fx.Annotate(NewPgxRWTransacter,
			fx.As(new(RWTransacter)),
			fx.ParamTags(``, ``, `name:"rw"`))),
	)
}
