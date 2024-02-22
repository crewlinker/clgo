package clbuildinfo_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/clbuildinfo"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestClbuildinfo(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clbuildinfo")
}

var _ = Describe("build info", func() {
	var info clbuildinfo.Info
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&info), clbuildinfo.TestProvide(), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should report version", func() {
		Expect(info).ToNot(BeNil())
		Expect(info.Version()).To(Equal("v0.0.0-test"))
	})
})
