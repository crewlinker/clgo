package clzap_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogging(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clzap")
}

var _ = Describe("regular logging", func() {
	var logs *zap.Logger
	var tmpfp string

	BeforeEach(func(ctx context.Context) {
		tmpfp = filepath.Join(os.TempDir(), fmt.Sprintf("test_logging_%d.log", time.Now().UnixNano()))
		app := fx.New(clzap.Fx(), clzap.Provide(), fx.Populate(&logs),
			fx.Decorate(func(cfg clzap.Config) clzap.Config {
				cfg.Outputs = []string{tmpfp}
				cfg.FxLevel = zapcore.InfoLevel

				return cfg
			}))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			Expect(app.Stop(ctx)).To(Succeed())
			Expect(os.Remove(tmpfp)).To(Succeed())
		})
	})

	It("should observe regular logging", func() {
		data, err := os.ReadFile(tmpfp)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(ContainSubstring("provided"))

		By("checking that by default follows lambda format")
		Expect(data).To(ContainSubstring(`,"message":`))
		Expect(data).To(ContainSubstring(`,"timestamp":`))
	})
})

var _ = Describe("test logging", func() {
	var logs *zap.Logger
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(clzap.TestProvide(), fx.Populate(&logs, &obs))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should not observe fx logging", func() {
		Expect(obs.FilterMessageSnippet("provided").Len()).To(BeNumerically("==", 0))
	})
})
