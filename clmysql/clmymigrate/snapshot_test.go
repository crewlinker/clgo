package clmymigrate_test

import (
	"context"
	"database/sql"
	"path/filepath"

	"github.com/crewlinker/clgo/clmysql"
	"github.com/crewlinker/clgo/clmysql/clmymigrate"
	"github.com/crewlinker/clgo/clzap"
	"github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"go.uber.org/fx"
)

var snapshot1 = filepath.Join("test_data", "snapshot", "snap1.sql")

var _ = Describe("snapshot migrater", func() {
	var sqldb *sql.DB
	var dbcfg *mysql.Config
	var mig clmysql.Migrater

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&sqldb, &dbcfg, &mig),
			clzap.TestProvide(),
			clmysql.TestProvide(),
			clmymigrate.SnapshotMigrated(snapshot1),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			// assert that the temp database is cleaned up.
			Expect(sql.OpenDB(lo.Must(mysql.NewConnector(dbcfg))).PingContext(ctx)).To(MatchError(MatchRegexp(`Unknown database`)))
		})

		DeferCleanup(app.Stop)
	})

	It("should create temp db and allow insert in migrated table", func(ctx context.Context) {
		Expect(mig).ToNot(BeNil())
		Expect(dbcfg.DBName).To(HavePrefix("temp_"))
		_, err := sqldb.ExecContext(ctx, "insert into profiles (id) values (1);")
		Expect(err).To(Succeed())
	})

	for range 10 {
		It("should run in a isolated database", func(ctx context.Context) {
			_, err := sqldb.ExecContext(ctx, "insert into profiles (id) values (1);")
			Expect(err).To(Succeed()) // fails when not isolated because id is unique
		})
	}
})
