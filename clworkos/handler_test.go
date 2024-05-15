package clworkos_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/clworkos"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestClworkos(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clworkos")
}

var _ = Describe("handler", func() {
	var hdlr *clworkos.Handler
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&hdlr), Provide(0))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(hdlr).NotTo(BeNil())
	})

	// @TODO test error handler endpoint
})

func Provide(clockAt int64) fx.Option {
	return fx.Options(
		clworkos.TestProvide(GinkgoTB(), clockAt),
		clzap.TestProvide(),
	)
}
