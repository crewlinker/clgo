package clpgmigrate_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgres(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clpostgres/clpgmigrate")
}

var _ = BeforeSuite(func() {
	Expect(godotenv.Load(filepath.Join("..", "..", "test.env"))).To(Succeed())
})
