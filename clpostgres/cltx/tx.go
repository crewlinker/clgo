// Package cltx is a small package for adding and retrieving pgx transactions from the context.
package cltx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// scope contetvalues.
type ctxKey string

// Tx returns the request scoped transaction initialized in the RWTransacter or
// ROTransacter middleware. If this was not done the function will panic.
func Tx(ctx context.Context) pgx.Tx {
	v := ctx.Value(ctxKey("tx"))
	if v == nil {
		panic("cltx: no tx in context")
	}

	tx, _ := v.(pgx.Tx)

	return tx
}

// WithTx returns a context with the provided tx added.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKey("tx"), tx)
}
