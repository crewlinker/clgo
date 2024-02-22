package clhealth_test

import (
	"context"
	"io"
	"testing"

	"github.com/crewlinker/clgo/clhealth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClhealth(t *testing.T) {
	t.Parallel()

	RegisterFailHandler(Fail)
	RunSpecs(t, "clhealth")
}

var _ = Describe("check health", func() {
	It("should check succeed", func(ctx context.Context) {
		code := 100

		clhealth.CheckHealthAndExit(ctx, io.Discard,
			[]string{"", "healthcheck", "https://google.com", "200"}, func(i int) { code = i })

		Expect(code).To(Equal(0))
	})

	It("invalid should cause 2", func(ctx context.Context) {
		code := 100

		clhealth.CheckHealthAndExit(ctx, io.Discard,
			[]string{"", "healthcheck", "https://google", "200"}, func(i int) { code = i })

		Expect(code).To(Equal(2))
	})

	It("invalid should cause 2", func(ctx context.Context) {
		code := 100

		clhealth.CheckHealthAndExit(ctx, io.Discard,
			[]string{"", "healthcheck", "https://google.com", "999"}, func(i int) { code = i })

		Expect(code).To(Equal(1))
	})
})
