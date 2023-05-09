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
	var htc *http.Client
	Describe("with tracing", func() {
		var tobs *tracetest.InMemoryExporter
		var trp *sdktrace.TracerProvider

		BeforeEach(func(ctx context.Context) {
			app := fx.New(fx.Populate(&tobs, &trp), clzap.Test(), clotel.Test())
			Expect(app.Start(ctx)).To(Succeed())
			DeferCleanup(app.Stop)
			htc = &http.Client{}
		})

		It("should construct and trace", func(ctx context.Context) {
			ctx, span := trp.Tracer("foo").Start(ctx, "my.span")
			defer span.End()

			Expect(htc).ToNot(BeNil())
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://google.com", nil)

			resp, err := htc.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			Expect(trp.ForceFlush(ctx)).To(Succeed())
			spans := tobs.GetSpans().Snapshots()
			Expect(spans).To(HaveLen(1))
			Expect(spans[0].Name()).To(Equal("HTTP GET"))
		})

		It("trace for default client", func(ctx context.Context) {
			ctx, span := trp.Tracer("foo").Start(ctx, "my.span")
			defer span.End()
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://google.com", nil)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			Expect(trp.ForceFlush(ctx)).To(Succeed())
			spans := tobs.GetSpans().Snapshots()
			Expect(spans).To(HaveLen(1))
			Expect(spans[0].Name()).To(Equal("HTTP GET"))
		})
	})
})
