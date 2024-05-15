package clworkos_test

import (
	"context"

	"github.com/crewlinker/clgo/clworkos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("handler", func() {
	var engine *clworkos.Engine
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&engine), Provide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do DI", func() {
		Expect(engine).NotTo(BeNil())
	})
})
