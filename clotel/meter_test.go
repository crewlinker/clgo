package clotel_test

import (
	"context"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/fx"
)

var _ = Describe("otel metrics", func() {
	var mp *sdkmetric.MeterProvider
	var mpi metric.MeterProvider
	var mr sdkmetric.Reader
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&mp, &mpi, &mr), clotel.Test, clzap.Test)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should provide metering", func(ctx context.Context) {
		ctr, err := mp.Meter("some_test").SyncInt64().Counter("some.counter")
		Expect(err).ToNot(HaveOccurred())
		ctr.Add(ctx, 100)
		ctr.Add(ctx, 10)

		metrics, err := mr.Collect(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(metrics.ScopeMetrics[0].Scope.Name).To(Equal("some_test"))
		Expect(metrics.ScopeMetrics[0].Metrics[0].Name).To(Equal("some.counter"))
		Expect(metrics.ScopeMetrics[0].Metrics[0].Data.(metricdata.Sum[int64]).DataPoints[0].Value).To(Equal(int64(110)))
	})
})
