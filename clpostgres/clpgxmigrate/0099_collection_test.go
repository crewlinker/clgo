package clpgxmigrate_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/clpostgres/clpgxmigrate"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/gomega"
)

func TestDefaultCollectionRegister(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, g := Setup(t)

		clpgxmigrate.Register(clpgxmigrate.NewStep(func(context.Context, pgx.Tx) error { return nil }))
		_, ok := clpgxmigrate.DefaulCollection.Step(99)
		g.Expect(ok).To(BeTrue())
	})
}
