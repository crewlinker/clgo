package clory_test

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clory"
	"github.com/crewlinker/clgo/clzap"
	orysdk "github.com/ory/client-go"
	mock "github.com/stretchr/testify/mock"
	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rpcv1 "github.com/crewlinker/clgo/clconnect/v1"
)

var _ = Describe("interceptors", func() {
	var ory *clory.Ory
	var front *MockFrontendAPI

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&ory),
			fx.Decorate(func(c clory.Config) clory.Config {
				c.PublicRPCProcedures = map[string]bool{"/clconnect.v1.ReadWriteService/CheckHealth": true}

				return c
			}),
			clory.Provide(),
			clzap.TestProvide(),
			WithMocked(&front))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should correctly return if procedure is public", func() {
		Expect(ory.IsPublicRPCMethod(connect.Spec{Procedure: "/clconnect.v1.ReadWriteService/Bogus"})).To(BeFalse())
		Expect(ory.IsPublicRPCMethod(connect.Spec{Procedure: "/clconnect.v1.ReadWriteService/CheckHealth"})).To(BeTrue())
	})

	It("should unauthenticate", func(ctx context.Context) {
		front.EXPECT().ToSession(mock.Anything).Return(orysdk.FrontendAPIToSessionRequest{})
		front.EXPECT().ToSessionExecute(mock.Anything).Return(nil, nil, errors.New("some error"))

		sess, err := callTestIntercept(ctx, ory)
		Expect(err).To(MatchError(clory.ErrUnauthenticated))
		Expect(sess).To(BeNil())
	})

	It("should authenticate", func(ctx context.Context) {
		front.EXPECT().ToSession(mock.Anything).Return(orysdk.FrontendAPIToSessionRequest{})
		front.EXPECT().ToSessionExecute(mock.Anything).Return(&orysdk.Session{Active: orysdk.PtrBool(true)}, nil, nil)

		sess, err := callTestIntercept(ctx, ory)
		Expect(err).ToNot(HaveOccurred())
		Expect(*sess.Active).To(BeTrue())
	})
})

func callTestIntercept(ctx context.Context, ory *clory.Ory) (*orysdk.Session, error) {
	var sess *orysdk.Session

	_, err := ory.Interceptor().WrapUnary(func(ctx context.Context,
		ar connect.AnyRequest,
	) (connect.AnyResponse, error) {
		sess = clory.Session(ctx)

		return connect.NewResponse(&rpcv1.CheckHealthResponse{}), nil
	})(ctx, connect.NewRequest(&rpcv1.CheckHealthRequest{}))

	return sess, err
}
