package clpostgres_test

import (
	"context"
	"database/sql"

	"ariga.io/atlas/sql/migrate"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("migrater", func() {
	var db *sql.DB
	var dbcfg *pgxpool.Config
	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&db, &dbcfg),
			fx.Provide(func() (migrate.Dir, error) { return migrate.NewLocalDir("test_data") }),
			clzap.Test, clpostgres.Test)
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			Expect(stdlib.OpenDB(*dbcfg.ConnConfig).PingContext(ctx).Error()).To(MatchRegexp(`database .* does not exist`))
		})

		DeferCleanup(app.Stop)
	})

	It("should create temp db and allow insert in migrated table", func(ctx context.Context) {
		Expect(dbcfg.ConnConfig.Database).To(HavePrefix("temp_"))
		_, err := db.ExecContext(ctx, "insert into profiles (id) values (1);")
		Expect(err).To(Succeed())
	})

	for i := 0; i < 10; i++ {
		It("should run in a isolated database", func(ctx context.Context) {
			_, err := db.ExecContext(ctx, "insert into profiles (id) values (1);")
			Expect(err).To(Succeed()) // fails when not isolated because id is unique
		})
	}
})
