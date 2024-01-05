package clconnect_test

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clconnect"
	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/crewlinker/clgo/clpostgres/cltx"
)

// ReadWrite represents the read-write side of the rpc.
type ReadWrite struct{}

// NewReadWrite inits the read-write rpc handler.
func NewReadWrite() (
	clconnectv1connect.ReadWriteServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadWriteServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadWriteServiceClient],
) {
	return ReadWrite{},
		clconnectv1connect.NewReadWriteServiceHandler,
		clconnectv1connect.NewReadWriteServiceClient
}

// ReadOnly represents the read-write side of the rpc.
type ReadOnly struct{}

// NewReadOnly inits the read-write rpc handler.
func NewReadOnly() (
	clconnectv1connect.ReadOnlyServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadOnlyServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadOnlyServiceClient],
) {
	return ReadOnly{},
		clconnectv1connect.NewReadOnlyServiceHandler,
		clconnectv1connect.NewReadOnlyServiceClient
}

// ErrInducedServerError is returned when the client on-purposed asked for a server error.
var ErrInducedServerError = errors.New("induced server error")

// CheckHealth implements the RPC method.
func (rw ReadWrite) CheckHealth(
	ctx context.Context, req *connect.Request[clconnectv1.CheckHealthRequest],
) (*connect.Response[clconnectv1.CheckHealthResponse], error) {
	if _, err := cltx.Tx(ctx).Exec(ctx, `UPDATE pg_catalog.pg_class SET relname = relname WHERE oid = -1;`); err != nil {
		return nil, fmt.Errorf("failed to exec sql: %w", err)
	}

	switch req.Msg.GetInduceError() {
	case clconnectv1.InducedError_INDUCED_ERROR_PANIC:
		panic("induced panic")
	case clconnectv1.InducedError_INDUCED_ERROR_UNKNOWN:
		return nil, ErrInducedServerError
	case clconnectv1.InducedError_INDUCED_ERROR_UNSPECIFIED:
		fallthrough
	default:
		return &connect.Response[clconnectv1.CheckHealthResponse]{
			Msg: &clconnectv1.CheckHealthResponse{
				Echo: req.Msg.GetEcho(),
			},
		}, nil
	}
}

// Foo implements the RPC method.
func (rw ReadOnly) Foo(
	ctx context.Context, req *connect.Request[clconnectv1.FooRequest],
) (*connect.Response[clconnectv1.FooResponse], error) {
	tx := cltx.Tx(ctx)
	if _, err := tx.Exec(ctx, `UPDATE pg_catalog.pg_class SET relname = relname WHERE oid = -1;`); err == nil {
		return nil, errors.New("should fail because read-only") //nolint:goerr113
	}

	return &connect.Response[clconnectv1.FooResponse]{}, nil
}
