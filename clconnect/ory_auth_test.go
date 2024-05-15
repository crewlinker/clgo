package clconnect_test

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clconnect"
	"github.com/crewlinker/clgo/clory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orysdk "github.com/ory/client-go"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"

	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"

	clconnectmock "github.com/crewlinker/clgo/clconnect/clconnectmock"
)

var _ = Describe("ory auth", func() {
	var hdl http.Handler
	var inj *clconnect.OryAuth
	var rwc clconnectv1connect.ReadWriteServiceClient
	var roc clconnectv1connect.ReadOnlyServiceClient
	var mory *clconnectmock.MockOry

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(fx.Annotate(&hdl, fx.ParamTags(`name:"clconnect"`)), &inj, &rwc, &roc),
			fx.Decorate(func(c clconnect.Config) clconnect.Config {
				c.PublicRPCProcedures = map[string]bool{"/clconnect.v1.ReadOnlyService/Foo": true}

				return c
			}),
			clconnect.ProvideOryAuth(),
			clory.Provide(),
			Provide(),
			fx.Provide(NewOryAuthReadOnly, NewOryAuthReadWrite),
			WithMockedOry(&mory),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do di", func() {
		Expect(hdl).ToNot(BeNil())
		Expect(inj).ToNot(BeNil())
		Expect(mory).ToNot(BeNil())
	})

	It("should authenticate", func(ctx context.Context) {
		mory.EXPECT().Authenticate(mock.Anything, mock.Anything, false).Return(orysdk.NewSession("foo"), nil)

		sess, err := callTestIntercept(ctx, inj)
		Expect(err).ToNot(HaveOccurred())
		Expect(sess).ToNot(BeNil())
	})

	It("should unauthenticate", func(ctx context.Context) {
		mory.EXPECT().Authenticate(mock.Anything, mock.Anything, false).Return(nil, clory.ErrUnauthenticated)

		sess, err := callTestIntercept(ctx, inj)
		Expect(err).To(MatchError(clory.ErrUnauthenticated))
		Expect(sess).To(BeNil())

		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodeUnauthenticated))
	})

	Describe("handler e2e", func() {
		It("should have an anonymous session", func(ctx context.Context) {
			mory.EXPECT().Authenticate(mock.Anything, mock.Anything, true).Return(clory.AnonymousSession, nil)

			req := &connect.Request[clconnectv1.FooRequest]{Msg: &clconnectv1.FooRequest{}}
			resp, err := roc.Foo(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Msg.GetBar()).To(Equal(clory.AnonymousSessionID))
		})
	})
})

// WithMockedOry is a test helper that mocks handler dependencies.
func WithMockedOry(m1 **clconnectmock.MockOry) fx.Option {
	return fx.Options(
		fx.Decorate(func(clconnect.Ory) clconnect.Ory {
			mock := clconnectmock.NewMockOry(GinkgoT())
			*m1 = mock

			return mock
		}),
	)
}

// test utility.
func callTestIntercept(ctx context.Context, inj *clconnect.OryAuth) (*orysdk.Session, error) {
	var sess *orysdk.Session

	_, err := inj.WrapUnary(func(ctx context.Context,
		ar connect.AnyRequest,
	) (connect.AnyResponse, error) {
		sess = clory.Session(ctx)

		return connect.NewResponse(&clconnectv1.CheckHealthResponse{}), nil
	})(ctx, connect.NewRequest(&clconnectv1.CheckHealthRequest{}))

	return sess, err
}

// OryAuthReadWrite represents the read-write side of the rpc.
type OryAuthReadWrite struct{}

// NewOryAutheadWrite inits the read-write rpc handler.
func NewOryAuthReadWrite() (
	clconnectv1connect.ReadWriteServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadWriteServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadWriteServiceClient],
) {
	return OryAuthReadWrite{},
		clconnectv1connect.NewReadWriteServiceHandler,
		clconnectv1connect.NewReadWriteServiceClient
}

// OryAuthReadOnly represents the read-write side of the rpc.
type OryAuthReadOnly struct{}

// NewOryAuthReadOnly inits the read-write rpc handler.
func NewOryAuthReadOnly() (
	clconnectv1connect.ReadOnlyServiceHandler,
	clconnect.ConstructHandler[clconnectv1connect.ReadOnlyServiceHandler],
	clconnect.ConstructClient[clconnectv1connect.ReadOnlyServiceClient],
) {
	return OryAuthReadOnly{},
		clconnectv1connect.NewReadOnlyServiceHandler,
		clconnectv1connect.NewReadOnlyServiceClient
}

// CheckHealth implements the RPC method.
func (rw OryAuthReadWrite) CheckHealth(
	ctx context.Context, req *connect.Request[clconnectv1.CheckHealthRequest],
) (*connect.Response[clconnectv1.CheckHealthResponse], error) {
	return &connect.Response[clconnectv1.CheckHealthResponse]{}, nil
}

// Foo implements the RPC method.
func (rw OryAuthReadOnly) Foo(
	ctx context.Context, req *connect.Request[clconnectv1.FooRequest],
) (*connect.Response[clconnectv1.FooResponse], error) {
	sess := clory.Session(ctx)

	return &connect.Response[clconnectv1.FooResponse]{
		Msg: &clconnectv1.FooResponse{Bar: sess.Id},
	}, nil
}
