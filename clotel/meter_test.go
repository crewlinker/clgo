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
	var mtp *sdkmetric.MeterProvider
	var mpi metric.MeterProvider
	var mtr sdkmetric.Reader
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&mtp, &mpi, &mtr), clotel.TestProvide(), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should provide metering", func(ctx context.Context) {
		ctr, err := mtp.Meter("some_test").Int64Counter("some.counter")
		Expect(err).ToNot(HaveOccurred())
		ctr.Add(ctx, 100)
		ctr.Add(ctx, 10)

		mrm := metricdata.ResourceMetrics{}
		err = mtr.Collect(ctx, &mrm)
		Expect(err).ToNot(HaveOccurred())

		Expect(mrm.ScopeMetrics[0].Scope.Name).To(Equal("some_test"))
		Expect(mrm.ScopeMetrics[0].Metrics[0].Name).To(Equal("some.counter"))
		sum, _ := mrm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Sum[int64])
		Expect(sum.DataPoints[0].Value).To(Equal(int64(110)))
	})
})
