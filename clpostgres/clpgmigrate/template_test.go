package clpgmigrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clpostgres/clpgmigrate"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var testTemplateDatabaseName clpgmigrate.TemplateDatabaseName

var _ = SynchronizedBeforeSuite(func(ctx context.Context) []byte {
	bootstrapDB, templateDB := os.Getenv("CLPOSTGRES_DATABASE_NAME"), "clgo_migrate"
	host, port, username, password := "localhost",
		os.Getenv("CLPOSTGRES_PORT"),
		os.Getenv("CLPOSTGRES_USERNAME"),
		os.Getenv("CLPOSTGRES_PASSWORD")
	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s",
		username, password, net.JoinHostPort(host, port), bootstrapDB)

	Expect(clpgmigrate.SetupTemplateDatabaseFromSnapshot(ctx, connString, templateDB, snapshot1)).To(Succeed())

	return []byte(templateDB)
}, func(b []byte) {
	testTemplateDatabaseName = clpgmigrate.TemplateDatabaseName(b)
})

var _ = SynchronizedAfterSuite(func(context.Context) {}, func(ctx context.Context) {
	bootstrapDB := os.Getenv("CLPOSTGRES_DATABASE_NAME")
	host, port, username, password := "localhost",
		os.Getenv("CLPOSTGRES_PORT"),
		os.Getenv("CLPOSTGRES_USERNAME"),
		os.Getenv("CLPOSTGRES_PASSWORD")
	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s",
		username, password, net.JoinHostPort(host, port), bootstrapDB)

	Expect(clpgmigrate.TeardownTemplateDatabase(ctx, connString, string(testTemplateDatabaseName))).To(Succeed())
})

var _ = Describe("template migrater", func() {
	var sqldb *sql.DB
	var dbcfg *pgxpool.Config
	var mig clpostgres.Migrater

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&sqldb, &dbcfg, &mig),
			clzap.TestProvide(),
			clpostgres.TestProvide(),
			clpgmigrate.TemplateMigrated(testTemplateDatabaseName),
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
