package clpgmigrate_test

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"

	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clpostgres/clpgmigrate"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/samber/lo"
	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed test_data/goose_migrations/*.sql
var gooseMigrations embed.FS

var _ = Describe("goose migrater", func() {
	var sqldb *sql.DB
	var dbcfg *pgxpool.Config
	var mig clpostgres.Migrater

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&sqldb, &dbcfg, &mig),
			clzap.TestProvide(),
			clpostgres.TestProvide(),
			clpgmigrate.GooseMigrated(lo.Must(fs.Sub(gooseMigrations, "test_data/goose_migrations"))),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(func(ctx context.Context) {
			Expect(stdlib.OpenDB(*dbcfg.ConnConfig).PingContext(ctx).Error()).To(MatchRegexp(`database .* does not exist`))
		})

		DeferCleanup(app.Stop)
	})

	It("should create temp db and allow insert in migrated table", func(ctx context.Context) {
		Expect(mig).ToNot(BeNil())
		Expect(dbcfg.ConnConfig.Database).To(HavePrefix("temp_"))

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
