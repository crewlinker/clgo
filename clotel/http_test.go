package clotel_test

import (
	"context"
	"net/http"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
)

var _ = Describe("client", func() {
	var hc *http.Client
	Describe("with tracing", func() {
		var tobs *tracetest.InMemoryExporter
		var tp *sdktrace.TracerProvider

		BeforeEach(func(ctx context.Context) {
			app := fx.New(fx.Populate(&tobs, &tp), clzap.Test(), clotel.Test)
			Expect(app.Start(ctx)).To(Succeed())
			DeferCleanup(app.Stop)
			hc = &http.Client{}
		})

		It("should construct and trace", func(ctx context.Context) {
			ctx, span := tp.Tracer("foo").Start(ctx, "my.span")
			defer span.End()

			Expect(hc).ToNot(BeNil())
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://google.com", nil)

			resp, err := hc.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			Expect(tp.ForceFlush(ctx)).To(Succeed())
			spans := tobs.GetSpans().Snapshots()
			Expect(spans).To(HaveLen(1))
			Expect(spans[0].Name()).To(Equal("HTTP GET"))
		})

		It("trace for default client", func(ctx context.Context) {
			ctx, span := tp.Tracer("foo").Start(ctx, "my.span")
			defer span.End()
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://google.com", nil)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			Expect(tp.ForceFlush(ctx)).To(Succeed())
			spans := tobs.GetSpans().Snapshots()
			Expect(spans).To(HaveLen(1))
			Expect(spans[0].Name()).To(Equal("HTTP GET"))
		})
	})
})
