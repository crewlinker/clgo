package clconnect_test

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"connectrpc.com/connect"
	"github.com/crewlinker/clgo/clauthn"
	"github.com/crewlinker/clgo/clauthz"
	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clconnect"
	clconnectv1 "github.com/crewlinker/clgo/clconnect/v1"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func TestClconnect(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clconnect")
}

var _ = BeforeSuite(func() {
	godotenv.Load(filepath.Join("..", "test.env"))
})

var _ = Describe("rpc", func() {
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

	It("should return health check on ok", func(ctx context.Context) {
		resp, err := rwc.CheckHealth(ctx,
			&connect.Request[clconnectv1.CheckHealthRequest]{Msg: &clconnectv1.CheckHealthRequest{Echo: "foo"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Msg.GetEcho()).To(Equal("foo"))
		Expect(obs.FilterMessage("handling request").All()).To(HaveLen(1))
	})

	It("should serve read-only", func(ctx context.Context) {
		_, err := roc.Foo(ctx, &connect.Request[clconnectv1.FooRequest]{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should trigger a server error", func(ctx context.Context) {
		_, err := rwc.CheckHealth(ctx,
			&connect.Request[clconnectv1.CheckHealthRequest]{Msg: &clconnectv1.CheckHealthRequest{
				InduceError: clconnectv1.InducedError_INDUCED_ERROR_UNKNOWN,
				Echo:        "foo",
			}})
		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodeUnknown))
		Expect(obs.FilterMessage("server error").All()).To(HaveLen(1))

		By("checking debug info")
		var debugInfo *errdetails.DebugInfo
		for _, detail := range cerr.Details() {
			val, err := detail.Value()
			Expect(err).ToNot(HaveOccurred())
			switch val := val.(type) {
			case *errdetails.DebugInfo:
				debugInfo = val
			}
		}

		Expect(debugInfo).ToNot(BeNil())
		Expect(len(debugInfo.GetStackEntries())).To(BeNumerically(">", 5))
	})

	It("should handle and log panics with stack traces", func(ctx context.Context) {
		_, err := rwc.CheckHealth(ctx,
			&connect.Request[clconnectv1.CheckHealthRequest]{Msg: &clconnectv1.CheckHealthRequest{
				InduceError: clconnectv1.InducedError_INDUCED_ERROR_PANIC,
				Echo:        "bar",
			}})

		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodeInternal))
		Expect(obs.FilterMessage("handling panic").All()).To(HaveLen(1))

		By("checking debug info")
		var debugInfo *errdetails.DebugInfo
		for _, detail := range cerr.Details() {
			val, err := detail.Value()
			Expect(err).ToNot(HaveOccurred())
			switch val := val.(type) {
			case *errdetails.DebugInfo:
				debugInfo = val
			}
		}

		Expect(debugInfo).ToNot(BeNil())
		Expect(len(debugInfo.GetStackEntries())).To(BeNumerically(">", 5))
	})

	It("should handle validation errors", func(ctx context.Context) {
		resp, err := rwc.CheckHealth(ctx, &connect.Request[clconnectv1.CheckHealthRequest]{
			Msg: &clconnectv1.CheckHealthRequest{Echo: ""},
		})
		Expect(resp).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(obs.FilterMessage("handling request").All()).To(BeEmpty())

		By("checking violation errors")
		var cerr *connect.Error
		Expect(errors.As(err, &cerr)).To(BeTrue())
		Expect(cerr.Code()).To(Equal(connect.CodeInvalidArgument))
		Expect(cerr.Details()).To(HaveLen(1))

		det, err := cerr.Details()[0].Value()
		Expect(err).ToNot(HaveOccurred())

		viol, ok := det.(*validate.Violations)
		Expect(ok).To(BeTrue())

		Expect(viol.GetViolations()).To(HaveLen(1))
		Expect(viol.GetViolations()[0].GetFieldPath()).To(Equal("echo"))
	})
})

func ProvidePgx() fx.Option {
	return fx.Options(
		Provide(),

		clconnect.ProvidePgxTransactors(),
		fx.Provide(NewReadOnly, NewReadWrite),
	)
}

func ProvideEnt() fx.Option {
	return fx.Options(
		Provide(),

		clconnect.ProvideEntTransactors[*modelTx, *modelClient](),
		fx.Supply(fx.Annotate(&modelClient{}, fx.ResultTags(`name:"rw"`))),
		fx.Supply(fx.Annotate(&modelClient{}, fx.ResultTags(`name:"ro"`))),
		fx.Provide(newEntReadOnly, newEntReadWrite),
	)
}

func Provide() fx.Option {
	return fx.Options(
		clauthn.TestProvide(),
		clauthz.TestProvide(clauthz.AllowAll()),
		claws.Provide(),
		clpostgres.TestProvide(),
		clzap.TestProvide(),

		clconnect.TestProvide[
			clconnectv1connect.ReadOnlyServiceHandler,
			clconnectv1connect.ReadWriteServiceHandler,
			clconnectv1connect.ReadOnlyServiceClient,
			clconnectv1connect.ReadWriteServiceClient,
		]("clconnect"),
	)
}
