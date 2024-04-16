package clenv_test

import (
	"context"
	"os"
	"testing"

	"github.com/crewlinker/clgo/clenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClenv(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clenv")
}

var _ = Describe("loading", Serial, func() {
	It("should load from git root", func(ctx context.Context) {
		Expect(clenv.LoadFromGitRoot(ctx, "clenv/testdata/foo.txt")).To(Succeed())

		Expect(os.Getenv("FOO")).To(Equal("bar"))
	})
})
