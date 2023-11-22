package clcdk_test

import (
	"testing"

	"github.com/crewlinker/clgo/clcdk"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClcdk(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clcdk")
}

var _ = Describe("scope", func() {
	It("should stringify scope names", func() {
		var name clcdk.ScopeName = "Foo"
		Expect(name.String()).To(Equal(`[Foo]`))
	})
})
