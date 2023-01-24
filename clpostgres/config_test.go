package clpostgres_test

import (
	"context"

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
				clzap.Test, clpostgres.Prod)
			Expect(app.Start(ctx)).To(Succeed())
			DeferCleanup(app.Stop)
		})

		It("should have build configs", func() {
			Expect(cfgs.ReadWrite.ConnConfig.Password).To(Equal("my-p&assword"))
			Expect(cfgs.ReadWrite.ConnConfig.Host).To(Equal("foo.read-write"))
			Expect(cfgs.ReadOnly.ConnConfig.Host).To(Equal("foo.read-only"))
		})
	})
})
