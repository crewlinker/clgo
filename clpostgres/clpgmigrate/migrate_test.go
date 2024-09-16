package clpgmigrate_test

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgres(t *testing.T) {
	godotenv.Load(filepath.Join("..", "..", "test.env"))

	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clpostgres/clpgmigrate")
}
