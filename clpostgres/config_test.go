package clpostgres_test

import (
	"context"
	"os"

	"github.com/crewlinker/clgo/claws"
	"github.com/crewlinker/clgo/clpostgres"
	"github.com/crewlinker/clgo/clzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var cfgs struct {
		fx.In
		ReadWrite *pgxpool.Config `name:"rw"`
		ReadOnly  *pgxpool.Config `name:"ro"`
	}

	Describe("non-iam config", func() {
		BeforeEach(func(ctx context.Context) {
			app := fx.New(
				fx.Populate(&cfgs),
				fx.Decorate(func(c clpostgres.Config) clpostgres.Config {
					c.Password = "my-p&assword"
					c.ReadOnlyHostname = "foo.read-only"
					c.ReadWriteHostname = "foo.read-write"

					return c
				}),
				clzap.TestProvide(), clpostgres.Provide())
			Expect(app.Start(ctx)).To(Succeed())
			DeferCleanup(app.Stop)
		})

		It("should have build configs", func() {
			Expect(cfgs.ReadWrite.ConnConfig.Password).To(Equal("my-p&assword"))
			Expect(cfgs.ReadWrite.ConnConfig.Host).To(Equal("foo.read-write"))
			Expect(cfgs.ReadOnly.ConnConfig.Host).To(Equal("foo.read-only"))
		})
	})
	Describe("iam config", Serial, func() {
		BeforeEach(func(ctx context.Context) {
			os.Setenv("AWS_ACCESS_KEY_ID", "a")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "b")
			os.Setenv("AWS_SESSION_TOKEN", "c")

			app := fx.New(
				fx.Populate(&cfgs),
				fx.Decorate(func(pgc clpostgres.Config) clpostgres.Config {
					pgc.IamAuthRegion = "eu-west-1"
					pgc.Password = "my-p&assword"
					pgc.ReadOnlyHostname = "foo.read-only"
					pgc.ReadWriteHostname = "foo.read-write"

					return pgc
				}),
				clzap.TestProvide(), claws.Provide(), clpostgres.Provide())
			Expect(app.Start(ctx)).To(Succeed())
			DeferCleanup(app.Stop)
		})

		It("should have build configs", func(ctx context.Context) {
			cfg := cfgs.ReadWrite.ConnConfig

			Expect(cfgs.ReadWrite.BeforeConnect(ctx, cfg)).To(Succeed())

			Expect(cfgs.ReadWrite.ConnConfig.Password).To(MatchRegexp(`^foo.read-write:([0-9]+)\?Action=connect`))
			Expect(cfgs.ReadWrite.ConnConfig.Host).To(Equal("foo.read-write"))
			Expect(cfgs.ReadOnly.ConnConfig.Host).To(Equal("foo.read-only"))
		})
	})

	Describe("conn string from env", Serial, func() {
		It("should give valid url with no env", func() {
			Expect(clpostgres.ConnStringFromEnvironment()).To(Equal(`postgres://postgres:postgres@localhost:5435/postgres?application_name=unknown.rw&sslmode=disable`))
		})

		It("should take envs into account", func() {
			os.Setenv("CLPOSTGRES_RW_HOSTNAME", "myrw")
			os.Setenv("CLPOSTGRES_APPLICATION_NAME", "app")
			defer os.Unsetenv("CLPOSTGRES_RW_HOSTNAME")
			defer os.Unsetenv("CLPOSTGRES_APPLICATION_NAME")

			Expect(clpostgres.ConnStringFromEnvironment()).To(Equal(`postgres://postgres:postgres@myrw:5435/postgres?application_name=app.rw&sslmode=disable`))
		})

		It("should take envs & kind into account", func() {
			os.Setenv("CLPOSTGRES_RW_HOSTNAME", "myrw")
			os.Setenv("CLPOSTGRES_RO_HOSTNAME", "myro")
			defer os.Unsetenv("CLPOSTGRES_RW_HOSTNAME")
			defer os.Unsetenv("CLPOSTGRES_RO_HOSTNAME")

			Expect(clpostgres.ConnStringFromEnvironment(clpostgres.ConfigKindReadOnly)).To(Equal(`postgres://postgres:postgres@myro:5435/postgres?application_name=unknown.ro&sslmode=disable`))
		})
	})
})
