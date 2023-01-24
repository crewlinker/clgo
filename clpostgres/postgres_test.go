package clpostgres_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "internal/clpostgres")
}

var _ = BeforeSuite(func() {
	Expect(godotenv.Load(filepath.Join("..", ".test.env")))
})

var _ = Describe("connect", func() {
	var obs *observer.ObservedLogs
	var pg struct {
		fx.In
		ReadWrite *sql.DB `name:"rw"`
		ReadOnly  *sql.DB `name:"ro"`
	}

	// provide it as a straight sql.db as well
	var scdb *sql.DB
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&pg, &obs, &scdb),
			clzap.Test, clpostgres.Test)
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
})
