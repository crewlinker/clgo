package clconfig_test

import (
	"context"
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/crewlinker/clgo/clconfig"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestClconfig(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clconfig")
}

var _ = Describe("config parsing", func() {
	type Conf1 struct {
		Foo string `env:"FOO"`
	}

	It("should parse correctly", func() {
		cfg1, err := clconfig.EnvConfigurer[Conf1]()(env.Options{Environment: map[string]string{"FOO": "bar"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg1.Foo).To(Equal("bar"))
	})

	It("should parse prefix", func() {
		cfg1, err := clconfig.EnvConfigurer[Conf1]("BAR_")(env.Options{Environment: map[string]string{"BAR_FOO": "bar"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg1.Foo).To(Equal("bar"))
	})

	It("should provide as dependency", Serial, func(ctx context.Context) {
		os.Setenv("FOO", "bar")
		DeferCleanup(os.Unsetenv, "FOO")

		var cfg1 Conf1
		app := fx.New(fx.Populate(&cfg1), clzap.Fx(), clzap.Test(), clconfig.Provide[Conf1]())
		Expect(app.Start(ctx)).To(Succeed())
		Expect(app.Stop(ctx)).To(Succeed())
		Expect(cfg1.Foo).To(Equal("bar"))
	})

	It("should allow supplying env options", func(ctx context.Context) {
		var cfg1 Conf1
		app := fx.New(
			fx.Supply(env.Options{Environment: map[string]string{"FOO": "bar2"}}),
			fx.Populate(&cfg1),
			clzap.Fx(),
			clzap.Test(),
			clconfig.Provide[Conf1]())

		Expect(app.Start(ctx)).To(Succeed())
		Expect(app.Stop(ctx)).To(Succeed())
		Expect(cfg1.Foo).To(Equal("bar2"))
	})
})
