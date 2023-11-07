package postgresresource_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/crewlinker/clgo/clcdk/postgresresource"
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgresResource(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "lambda/postgresresource")
}

var _ = BeforeSuite(func() {
	godotenv.Load(filepath.Join("..", "..", "test.env"))
})

var _ = Describe("full app dependencies", func() {
	It("should wire up all dependencies as in actual deployment", func(ctx context.Context) {
		os.Setenv("CLZAP_LEVEL", "panic")
		DeferCleanup(os.Unsetenv, "CLZAPP_LEVEL")

		Expect(fx.New(postgresresource.Prod("v0.0.1")).Start(ctx)).To(Succeed())
	})
})

// WithMocked is a test helper that mocks handler dependencies.
func WithMocked(msm **MockSecretsManager) fx.Option {
	return fx.Options(
		fx.Decorate(func(postgresresource.SecretsManager) postgresresource.SecretsManager {
			mock := NewMockSecretsManager(GinkgoT())
			*msm = mock

			return mock
		}),
	)
}
