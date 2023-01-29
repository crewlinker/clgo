package clpostgres_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clotel"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clpostgres")
}

var _ = BeforeSuite(func() {
	Expect(godotenv.Load(filepath.Join("..", ".test.env")))
})

var _ = Describe("connect", func() {
	var obs *observer.ObservedLogs
	var tobs *tracetest.InMemoryExporter
	var tp *trace.TracerProvider
	var pg struct {
		fx.In
		ReadWrite *sql.DB `name:"rw"`
		ReadOnly  *sql.DB `name:"ro"`
	}

	// provide it as a straight sql.db as well
	var scdb *sql.DB
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&pg, &obs, &scdb, &tobs, &tp),
			clzap.Test, clpostgres.Test, clotel.Test)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)

		DeferCleanup(func(ctx context.Context) {
			Expect(app.Stop(ctx)).To(Succeed())
			_, err := pg.ReadOnly.QueryContext(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError("sql: database is closed"))
			_, err = pg.ReadWrite.QueryContext(ctx, `SELECT * FROM foo`)
			Expect(err).To(MatchError("sql: database is closed"))
		})
	})

	It("should work without contextual logger/tracer", func(ctx context.Context) {
		Expect(scdb).To(Equal(pg.ReadWrite))
		Expect(pg.ReadWrite.PingContext(ctx))
		Expect(pg.ReadOnly.PingContext(ctx))
		Expect(obs.FilterMessage("Query").Len()).To(BeNumerically(">=", 4))
	})

	It("should have observed traces", func(ctx context.Context) {
		var num int
		Expect(pg.ReadWrite.QueryRowContext(ctx, "SELECT 42").Scan(&num)).To(Succeed())

		Expect(tp.ForceFlush(ctx)).To(Succeed())
		Expect(len(tobs.GetSpans().Snapshots())).To(BeNumerically(">", 4))
		Expect(string(tobs.GetSpans().Snapshots()[0].Attributes()[0].Key)).To(Equal("db.name"))
	})
})
