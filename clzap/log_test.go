package clzap_test

import (
	"bytes"
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

var _ = Describe("development logging", func() {
	var logs *zap.Logger
	var tmpfp string

	BeforeEach(func(ctx context.Context) {
		tmpfp = filepath.Join(os.TempDir(), fmt.Sprintf("test_logging2_%d.log", time.Now().UnixNano()))
		app := fx.New(clzap.Fx(), clzap.Provide(), fx.Populate(&logs),
			fx.Decorate(func(cfg clzap.Config) clzap.Config {
				cfg.Outputs = []string{tmpfp}
				cfg.FxLevel = zapcore.InfoLevel
				cfg.DevelopmentEncodingConfig = true
				cfg.ConsoleEncoding = true

				return cfg
			}))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			Expect(app.Stop(ctx)).To(Succeed())
			Expect(os.Remove(tmpfp)).To(Succeed())
		})
	})

	It("should see development useful logging", func() {
		data, err := os.ReadFile(tmpfp)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(ContainSubstring("provided"))
		Expect(data).To(ContainSubstring("\tINFO\t"))
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

var _ = Describe("second core", func() {
	var logs *zap.Logger
	var obs *observer.ObservedLogs
	var buf *bytes.Buffer

	BeforeEach(func(ctx context.Context) {
		buf = bytes.NewBuffer(nil)
		zc2 := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(buf), zapcore.DebugLevel)

		app := fx.New(clzap.TestProvide(), fx.Populate(&logs, &obs), fx.Supply(&clzap.SecondaryCore{
			Core: zc2, Name: "some_core_name",
		}))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)

		DeferCleanup(func() {
			Expect(buf.String()).To(ContainSubstring("log something for secondary core"))
			Expect(buf.String()).To(ContainSubstring("logger initialized with secondary core"))
			Expect(buf.String()).To(ContainSubstring("some_core_name"))
		})
	})

	It("should not observe fx logging", func() {
		logs.Info("log something for secondary core")
	})
})
