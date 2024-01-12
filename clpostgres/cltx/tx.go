// Package cltx is a small package for adding and retrieving pgx transactions from the context.
package cltx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// scope contetvalues.
type ctxKey string

// Pgx returns the request scoped transaction initialized in the RWTransacter or
// ROTransacter middleware. If this was not done the function will panic.
func Pgx(ctx context.Context) pgx.Tx {
	return Tx[pgx.Tx](ctx)
}

// WithPgx returns a context with the provided pgx tx added.
func WithPgx(ctx context.Context, tx pgx.Tx) context.Context {
	return WithTx(ctx, tx)
}

// WithTx returns a context with the provided tx added.
func WithTx[T any](ctx context.Context, tx T) context.Context {
	return context.WithValue(ctx, ctxKey("tx"), tx)
}

// Tx is a generic transaction reader to support various transaction
// types. E.g: Pgx, sql.Tx, Ent' model tx.
func Tx[T any](ctx context.Context) T {
	v := ctx.Value(ctxKey("tx"))
	if v == nil {
		panic("cltx: no tx in context")
	}

	tx, ok := v.(T)
	if !ok {
		var exp T

		panic(fmt.Sprintf("cltx: wrong tx type in context, got: %T, expected: %T", v, exp))
	}

	return tx
}
