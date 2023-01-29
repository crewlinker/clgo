package clotel_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
)

func TestClotel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clotel")
}

var _ = Describe("otel", func() {
	var tp *sdktrace.TracerProvider
	var tobs *tracetest.InMemoryExporter
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&tp, &tobs), clotel.Test, clzap.Test)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should provide tracing", func(ctx context.Context) {
		_, span := tp.Tracer("test").Start(ctx, "my-span")
		span.End()

		Expect(tp.ForceFlush(ctx)).To(Succeed())
		spans := tobs.GetSpans().Snapshots()

		Expect(spans).To(HaveLen(1))
		Expect(spans[0].Name()).To(Equal("my-span"))
	})
})
