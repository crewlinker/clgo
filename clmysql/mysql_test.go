package clmysql_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clmysql"
	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clzap"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
)

func TestMysql(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clmysql")
}

var _ = BeforeSuite(func() {
	godotenv.Load(filepath.Join("..", "test.env"))
})

var _ = Describe("plain", func() {
	var rw *sql.DB
	var rw2 *sql.DB
	var ro *sql.DB

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&rw),
			fx.Populate(fx.Annotate(&rw2, fx.ParamTags(`name:"my_rw"`))),
			fx.Populate(fx.Annotate(&ro, fx.ParamTags(`name:"my_ro"`))),
			clmysql.TestProvide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should have performed DI", func() {
		Expect(rw).ToNot(BeNil())
		Expect(rw2).ToNot(BeNil())
		Expect(ro).ToNot(BeNil())
	})
})

var _ = Describe("with otel", func() {
	var rw *sql.DB
	var tobs *tracetest.InMemoryExporter
	var trp *sdktrace.TracerProvider
	var mtr sdkmetric.Reader

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&rw, &tobs, &trp, &mtr),
			clotel.TestProvide(),
			clmysql.TestProvide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should have performed DI", func() {
		Expect(rw).ToNot(BeNil())
		Expect(tobs).ToNot(BeNil())
	})

	It("should observe traces", func(ctx context.Context) {
		var num int
		Expect(rw.QueryRowContext(ctx, "SELECT 42").Scan(&num)).To(Succeed())
		Expect(num).To(BeNumerically("==", 42))

		Expect(trp.ForceFlush(ctx)).To(Succeed())
		Expect(len(tobs.GetSpans().Snapshots())).To(BeNumerically(">", 2))
		Expect(string(tobs.GetSpans().Snapshots()[0].Attributes()[0].Key)).To(Equal("db.user"))

		rm := metricdata.ResourceMetrics{}
		err := mtr.Collect(ctx, &rm)
		Expect(err).ToNot(HaveOccurred())

		Expect(rm.ScopeMetrics[0].Scope.Name).To(Equal("github.com/XSAM/otelsql"))
	})
})
