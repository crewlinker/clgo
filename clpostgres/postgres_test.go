package clpostgres_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestPostgres(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clpostgres")
}

var _ = BeforeSuite(func() {
	Expect(godotenv.Load(filepath.Join("..", ".test.env"))).To(Succeed())
})

var _ = Describe("pgx pool", func() {
	var rwp, rop *pgxpool.Pool
	var pool *pgxpool.Pool
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&pool),
			fx.Populate(
				fx.Annotate(&rwp, fx.ParamTags(`name:"rw"`)),
				fx.Annotate(&rop, fx.ParamTags(`name:"ro"`))),
			clpostgres.MigratedTest("test_data", false),
			clzap.Test())
		Expect(app.Start(ctx)).To(Succeed())

		DeferCleanup(func(ctx context.Context) {
			Expect(app.Stop(ctx)).To(Succeed())
			_, err := rop.Query(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError(`closed pool`))
			_, err = rwp.Query(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError(`closed pool`))
		})
	})

	It("should provide rw/ro pool", func() {
		Expect(rwp).ToNot(BeNil())
		Expect(rop).ToNot(BeNil())
		Expect(pool).ToNot(BeNil())
	})
})

var _ = Describe("observe", func() {
	var logs *zap.Logger
	var obs *observer.ObservedLogs
	var mtr sdkmetric.Reader
	var tobs *tracetest.InMemoryExporter
	var trp *sdktrace.TracerProvider
	var pgs struct {
		fx.In
		ReadWrite *sql.DB `name:"rw"`
		ReadOnly  *sql.DB `name:"ro"`
	}

	// provide it as a straight sql.db as well
	var scdb *sql.DB
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&logs, &pgs, &obs, &scdb, &tobs, &trp, &mtr),
			clzap.Test(), clpostgres.MigratedTest("test_data", false), clotel.Test())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)

		DeferCleanup(func(ctx context.Context) {
			Expect(app.Stop(ctx)).To(Succeed())
			_, err := pgs.ReadOnly.QueryContext(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError("sql: database is closed"))
			_, err = pgs.ReadWrite.QueryContext(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError("sql: database is closed"))
		})
	})

	It("should work with contextual logger/tracer", func(ctx context.Context) {
		ctx = clzap.WithLogger(ctx, logs)
		ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: trace.TraceID{0x01},
			SpanID:  trace.SpanID{0x02},
		}))

		Expect(scdb).To(Equal(pgs.ReadWrite))

		By("no-op sql execution to trigger contextual logging")
		_, err := pgs.ReadWrite.ExecContext(ctx, `BEGIN;COMMIT;`)
		Expect(err).To(Succeed())
		_, err = pgs.ReadOnly.ExecContext(ctx, `BEGIN;COMMIT;`)
		Expect(err).To(Succeed())

		qlogs := obs.FilterMessage("Query")
		Expect(qlogs.Len()).To(BeNumerically(">=", 2))
		Expect(qlogs.All()[len(qlogs.All())-1].ContextMap()).To(
			HaveKeyWithValue("trace_id", "1-01000000-000000000000000000000000"))
	})

	It("should have observed traces and metrics", func(ctx context.Context) {
		var num int
		Expect(pgs.ReadWrite.QueryRowContext(ctx, "SELECT 42").Scan(&num)).To(Succeed())

		Expect(trp.ForceFlush(ctx)).To(Succeed())
		Expect(len(tobs.GetSpans().Snapshots())).To(BeNumerically(">", 4))
		Expect(string(tobs.GetSpans().Snapshots()[0].Attributes()[0].Key)).To(Equal("db.user"))

		rm := metricdata.ResourceMetrics{}
		err := mtr.Collect(ctx, &rm)
		Expect(err).ToNot(HaveOccurred())

		Expect(rm.ScopeMetrics[0].Scope.Name).To(Equal("github.com/XSAM/otelsql"))
	})
})
