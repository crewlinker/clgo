package clconnect_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clconnect"
	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/crewlinker/clgo/clpostgres/cltx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("pgx", func() {
	var hdl http.Handler
	var rwc clconnectv1connect.ReadWriteServiceClient
	var roc clconnectv1connect.ReadOnlyServiceClient
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(fx.Annotate(&hdl, fx.ParamTags(`name:"clconnect"`)), &rwc, &roc, &obs),
			ProvidePgx(),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should setup di", func() {
		Expect(hdl).ToNot(BeNil())
		Expect(rwc).ToNot(BeNil())
		Expect(roc).ToNot(BeNil())
	})

	It("should call read-only rpc", func(ctx context.Context) {
		_, err := roc.Foo(ctx, &connect.Request[clconnectv1.FooRequest]{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should call read-write rpc", func(ctx context.Context) {
		_, err := rwc.CheckHealth(ctx,
			&connect.Request[clconnectv1.CheckHealthRequest]{Msg: &clconnectv1.CheckHealthRequest{Echo: "foo"}})
		Expect(err).ToNot(HaveOccurred())
	})
})

// pgxReadWrite represents the read-write side of the rpc.
type pgxReadWrite struct{}

// NewReadWrite inits the read-write rpc handler.
func NewPgxReadWrite() (
	clconnectv1connect.ReadWriteServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadWriteServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadWriteServiceClient],
) {
	return pgxReadWrite{},
		clconnectv1connect.NewReadWriteServiceHandler,
		clconnectv1connect.NewReadWriteServiceClient
}

// pgxReadOnly represents the read-write side of the rpc.
type pgxReadOnly struct{}

// NewReadOnly inits the read-write rpc handler.
func NewPgxReadOnly() (
	clconnectv1connect.ReadOnlyServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadOnlyServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadOnlyServiceClient],
) {
	return pgxReadOnly{},
		clconnectv1connect.NewReadOnlyServiceHandler,
		clconnectv1connect.NewReadOnlyServiceClient
}

// CheckHealth implements the RPC method.
func (rw pgxReadWrite) CheckHealth(
	ctx context.Context, req *connect.Request[clconnectv1.CheckHealthRequest],
) (*connect.Response[clconnectv1.CheckHealthResponse], error) {
	if _, err := cltx.Pgx(ctx).Exec(ctx, `UPDATE pg_catalog.pg_class SET relname = relname WHERE oid = -1;`); err != nil {
		return nil, fmt.Errorf("failed to exec sql: %w", err)
	}

	return &connect.Response[clconnectv1.CheckHealthResponse]{}, nil
}

// Foo implements the RPC method.
func (rw pgxReadOnly) Foo(
	ctx context.Context, req *connect.Request[clconnectv1.FooRequest],
) (*connect.Response[clconnectv1.FooResponse], error) {
	tx := cltx.Pgx(ctx)
	if _, err := tx.Exec(ctx, `UPDATE pg_catalog.pg_class SET relname = relname WHERE oid = -1;`); err == nil {
		return nil, errors.New("should fail because read-only") //nolint:goerr113
	}

	return &connect.Response[clconnectv1.FooResponse]{}, nil
}
