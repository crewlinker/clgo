package clzap

import (
	"io"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/fx"
)

// Test is a convenient fx option setup that can easily be included in all tests. It observed the logs
// for assertion and writes console output to the GinkgoWriter so all logs can easily be inspected if
// tests fail.
func Test() fx.Option {
	return fx.Options(Fx(),
		// in tests, always provide the ginkgo writer as the output writer so failing tests immediately show
		// the complete console output.
		fx.Supply(fx.Annotate(ginkgo.GinkgoWriter, fx.As(new(io.Writer)))),
		Observed())
}
