package clpgxmigrate_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/crewlinker/clgo/clpostgres/clpgxmigrate"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/gomega"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	t.Run("read version fail", func(t *testing.T) {
		t.Parallel()
		ctx, g, prv := SetupProvider(t)

		_, err := prv.ReadSchemaVersion(ctx)
		g.Expect(err).To(MatchError(MatchRegexp(`relation.*does not exist`)))
	})

	t.Run("migrate and read version", func(t *testing.T) {
		t.Parallel()
		ctx, g, prv := SetupProvider(t)

		result, err := prv.Migrate(ctx, math.MaxInt64)
		g.Expect(err).To(Succeed())
		g.Expect(result.AppliedVersions).To(Equal([]int64{99}))

		status, err := prv.Status(ctx)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.CurrentVersion).To(Equal(int64(99)))
	})

	t.Run("migrate with nothing to apply", func(t *testing.T) {
		t.Parallel()
		ctx, g, prv := SetupProvider(t)

		result, err := prv.Migrate(ctx, 0)
		g.Expect(err).To(Succeed())
		g.Expect(result.AppliedVersions).To(BeNil())

		status, err := prv.Status(ctx)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status.CurrentVersion).To(Equal(int64(0)))
	})

	t.Run("migrate with faulty step", func(t *testing.T) {
		t.Parallel()
		coll := clpgxmigrate.NewCollection()
		coll.Register("004_foo", clpgxmigrate.NewStep(func(context.Context, pgx.Tx) error { return errors.New("fail") }))

		ctx, g, conn := SetupConn(t)
		prv := clpgxmigrate.NewProvider(conn, clpgxmigrate.WithCollection(coll))

		_, err := prv.Migrate(ctx, math.MaxInt64)
		g.Expect(err).To(MatchError(MatchRegexp("failed to apply")))
		g.Expect(err).To(MatchError(MatchRegexp("fail")))
	})
}
