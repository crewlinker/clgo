package clconnect_test

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/crewlinker/clgo/clauth"
	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clconnect"
	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clpostgres/cltx"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("ent", func() {
	var hdl http.Handler
	var rwc clconnectv1connect.ReadWriteServiceClient
	var roc clconnectv1connect.ReadOnlyServiceClient
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			// provide connect apis
			fx.Populate(fx.Annotate(&hdl, fx.ParamTags(`name:"clconnect"`)), &rwc, &roc, &obs),
			clconnect.TestProvide[
				clconnectv1connect.ReadOnlyServiceHandler,
				clconnectv1connect.ReadWriteServiceHandler,
				clconnectv1connect.ReadOnlyServiceClient,
				clconnectv1connect.ReadWriteServiceClient,
			]("clconnect"),

			// ent transactor
			clconnect.ProvideEntTransactors[*modelTx, *modelClient](),
			fx.Supply(fx.Annotate(&modelClient{}, fx.ResultTags(`name:"rw"`))),
			fx.Supply(fx.Annotate(&modelClient{}, fx.ResultTags(`name:"ro"`))),

			// general provides
			fx.Provide(newEntReadOnly, newEntReadWrite),
			clauth.TestProvide(map[string]string{}),
			claws.Provide(),
			clpostgres.TestProvide(),
			clzap.TestProvide())

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
		resp, err := rwc.CheckHealth(ctx,
			&connect.Request[clconnectv1.CheckHealthRequest]{Msg: &clconnectv1.CheckHealthRequest{Echo: "foo"}})
		Expect(err).ToNot(HaveOccurred())

		Expect(resp.Msg.GetEcho()).To(Equal("bar"))
	})
})

// test ent model Tx.
type modelTx struct{}

func (modelTx) Commit() error   { return nil }
func (modelTx) Rollback() error { return nil }
func (modelTx) Foo() string     { return "bar" }

// test Ent model client.
type modelClient struct{}

func (modelClient) BeginTx(context.Context, *entsql.TxOptions) (*modelTx, error) {
	return &modelTx{}, nil
}

// ReadWrite represents the read-write side of the rpc.
type entReadWrite struct{}

// NewReadWrite inits the read-write rpc handler.
func newEntReadWrite() (
	clconnectv1connect.ReadWriteServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadWriteServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadWriteServiceClient],
) {
	return entReadWrite{},
		clconnectv1connect.NewReadWriteServiceHandler,
		clconnectv1connect.NewReadWriteServiceClient
}

// ReadOnly represents the read-write side of the rpc.
type entReadOnly struct{}

// NewReadOnly inits the read-write rpc handler.
func newEntReadOnly() (
	clconnectv1connect.ReadOnlyServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadOnlyServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadOnlyServiceClient],
) {
	return entReadOnly{},
		clconnectv1connect.NewReadOnlyServiceHandler,
		clconnectv1connect.NewReadOnlyServiceClient
}

// CheckHealth implements the RPC method.
func (rw entReadWrite) CheckHealth(
	ctx context.Context, req *connect.Request[clconnectv1.CheckHealthRequest],
) (*connect.Response[clconnectv1.CheckHealthResponse], error) {
	tx := cltx.Tx[*modelTx](ctx)

	return &connect.Response[clconnectv1.CheckHealthResponse]{
		Msg: &clconnectv1.CheckHealthResponse{Echo: tx.Foo()},
	}, nil
}

// Foo implements the RPC method.
func (rw entReadOnly) Foo(
	ctx context.Context, req *connect.Request[clconnectv1.FooRequest],
) (*connect.Response[clconnectv1.FooResponse], error) {
	tx := cltx.Tx[*modelTx](ctx)
	if tx == nil {
		panic("must have tx")
	}

	return &connect.Response[clconnectv1.FooResponse]{}, nil
}
