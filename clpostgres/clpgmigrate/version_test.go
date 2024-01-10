package clpgmigrate_test

import (
	"context"
	"database/sql"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clpostgres/clpgmigrate"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"go.uber.org/zap/zaptest/observer"
)

var _ = Describe("version migrater", func() {
	var sqldb *sql.DB
	var dbcfg *pgxpool.Config
	var versmig clpostgres.Migrater
	var obs *observer.ObservedLogs

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&sqldb, &dbcfg, &versmig, &obs),
			fx.Provide(func() (migrate.Dir, error) { return sqltool.NewGolangMigrateDir("test_data") }),
			clzap.TestProvide(),
			clpostgres.TestProvide(),
			clpgmigrate.VersionMigrated(true))
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			Expect(stdlib.OpenDB(*dbcfg.ConnConfig).PingContext(ctx).Error()).To(MatchRegexp(`database .* does not exist`))
		})

		DeferCleanup(app.Stop)
	})

	It("should create temp db and allow insert in migrated table", func(ctx context.Context) {
		Expect(dbcfg.ConnConfig.Database).To(HavePrefix("temp_"))
		_, err := sqldb.ExecContext(ctx, "insert into profiles (id) values (1);")
		Expect(err).To(Succeed())

		By("checking that sql has been logged")
		Expect(obs.FilterMessage("statement").All()[0].ContextMap()["sql"]).To(ContainSubstring(`CREATE TABLE`))
	})

	for i := 0; i < 10; i++ {
		It("should run in a isolated database", func(ctx context.Context) {
			_, err := sqldb.ExecContext(ctx, "insert into profiles (id) values (1);")
			Expect(err).To(Succeed()) // fails when not isolated because id is unique
		})
	}
})
