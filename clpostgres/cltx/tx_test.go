package cltx_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clpostgres/cltx"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestCltx(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clpostgres/cltx")
}

var _ = BeforeSuite(func() {
	godotenv.Load(filepath.Join("..", "..", "test.env"))
})

var _ = Describe("tx", func() {
	var db *pgxpool.Pool

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&db),
			clpostgres.TestProvide(),
			clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should panic without tx", func(ctx context.Context) {
		Expect(func() { cltx.Tx(ctx) }).To(Panic())
	})

	It("should add an remove tx from ctx", func(ctx context.Context) {
		tx1, err := db.Begin(ctx)
		Expect(err).ToNot(HaveOccurred())
		defer tx1.Rollback(ctx)

		ctx = cltx.WithTx(ctx, tx1)
		tx2 := cltx.Tx(ctx)

		Expect(tx2).To(Equal(tx1))
	})
})
