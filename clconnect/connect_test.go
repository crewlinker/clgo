package clconnect_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clconnect"
	"github.com/crewlinker/clgo/clconnect/v1/clconnectv1connect"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestClconnect(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clconnect")
}

var _ = Describe("rpc", func() {
	var hdl http.Handler
	var rwc clconnectv1connect.ReadWriteServiceClient
	var roc clconnectv1connect.ReadOnlyServiceClient

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(fx.Annotate(&hdl, fx.ParamTags(`name:"clconnect"`)), &rwc, &roc),
			clconnect.TestProvide[
				clconnectv1connect.ReadOnlyServiceHandler,
				clconnectv1connect.ReadWriteServiceHandler,
				clconnectv1connect.ReadOnlyServiceClient,
				clconnectv1connect.ReadWriteServiceClient,
			]("clconnect"),

			fx.Provide(NewReadOnly, NewReadWrite),
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
})

//
// Test Implementation
//

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
