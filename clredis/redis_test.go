package clredis_test

import (
	"context"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clredis"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
)

func TestClredis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clredis")
}

var _ = Describe("redis", func() {
	var red redis.UniversalClient
	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&red),
			fx.Decorate(func(c clredis.Config) clredis.Config {
				c.Addrs = []string{"localhost:6378"} // use our docker-hosted redis
				return c
			}), clredis.Test, clzap.Test())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should have a functioning", func(ctx context.Context) {
		Expect(red.Ping(ctx).Err()).To(Succeed())
	})
})

var _ = Describe("redis observed", func() {
	var red redis.UniversalClient
	var mr sdkmetric.Reader
	var tobs *tracetest.InMemoryExporter
	var tp *sdktrace.TracerProvider

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&red, &tobs, &tp, &mr),
			fx.Decorate(func(c clredis.Config) clredis.Config {
				c.Addrs = []string{"localhost:6378"} // use our docker-hosted redis
				return c
			}), clredis.Test, clzap.Test(), clotel.Test)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should trace and measure redis client interactions", func(ctx context.Context) {
		ctx, _ = tp.Tracer("tests").Start(ctx, "my-test")

		_, err := red.Set(ctx, "foo", "bar", time.Second).Result()
		Expect(err).To(Succeed())

		By("checking traces")
		Expect(tp.ForceFlush(ctx)).To(Succeed())
		Expect(len(tobs.GetSpans().Snapshots())).To(BeNumerically(">", 0))
		Expect(string(tobs.GetSpans().Snapshots()[0].Attributes()[0].Key)).To(Equal("db.system"))

		By("checking metrics")
		rm := metricdata.ResourceMetrics{}
		err = mr.Collect(ctx, &rm)
		Expect(err).ToNot(HaveOccurred())
		Expect(rm.ScopeMetrics[0].Scope.Name).To(Equal("github.com/redis/go-redis/extra/redisotel"))
	})
})
