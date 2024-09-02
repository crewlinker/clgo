package clpgxmigrate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/crewlinker/clgo/clpostgres/clpgxmigrate"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/gomega"
)

func TestCollection(t *testing.T) {
	t.Parallel()

	for idx, tcase := range []struct {
		filename string
		expErr   OmegaMatcher
	}{
		{"foo", MatchError(MatchRegexp(`no .* separator`))},
		{"foo_ab", MatchError(MatchRegexp(`invalid syntax`))},
		{"0000_ab", MatchError(MatchRegexp(`is not greater then zero`))},
	} {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			t.Parallel()
			_, g := Setup(t)

			coll := clpgxmigrate.NewCollection()
			err := coll.Register(tcase.filename, nil)
			g.Expect(err).To(HaveOccurred())

			g.Expect(err).To(tcase.expErr)
		})
	}

	t.Run("already registered", func(t *testing.T) {
		t.Parallel()
		_, g := Setup(t)

		coll := clpgxmigrate.NewCollection()
		g.Expect(coll.Register("001_foo", nil)).To(Succeed())
		g.Expect(coll.Register("001_foo", nil)).To(MatchError(MatchRegexp(`already registered`)))
	})

	t.Run("versions", func(t *testing.T) {
		t.Parallel()

		for idx, tcase := range []struct {
			coll        func() clpgxmigrate.Collection
			expVersions OmegaMatcher
		}{
			{
				coll: func() clpgxmigrate.Collection {
					return clpgxmigrate.NewCollection()
				},
				expVersions: BeEmpty(),
			},
			{
				coll: func() clpgxmigrate.Collection {
					coll := clpgxmigrate.NewCollection()
					_ = coll.Register("001_foo", nil)

					return coll
				},
				expVersions: Equal([]int64{1}),
			},
			{
				coll: func() clpgxmigrate.Collection {
					coll := clpgxmigrate.NewCollection()
					_ = coll.Register("001_foo", nil)
					_ = coll.Register("001_foo", nil)

					return coll
				},
				expVersions: Equal([]int64{1}),
			},
			{
				coll: func() clpgxmigrate.Collection {
					coll := clpgxmigrate.NewCollection()
					_ = coll.Register("009_foo", nil)
					_ = coll.Register("001_foo", nil)

					return coll
				},
				expVersions: Equal([]int64{1, 9}),
			},
		} {
			t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
				t.Parallel()
				_, g := Setup(t)
				g.Expect(tcase.coll().Versions()).To(tcase.expVersions)
			})
		}
	})

	t.Run("get step", func(t *testing.T) {
		t.Parallel()
		_, g := Setup(t)
		step1 := clpgxmigrate.NewStep(func(context.Context, pgx.Tx) error { return nil })

		coll := clpgxmigrate.NewCollection()
		g.Expect(coll.Register("0001_step", step1)).To(Succeed())

		step2, ok := coll.Step(1)
		g.Expect(ok).To(BeTrue())
		g.Expect(fmt.Sprint(step2)).To(Equal(fmt.Sprint(step1)))

		step3, ok := coll.Step(999)
		g.Expect(ok).To(BeFalse())
		g.Expect(step3).To(BeNil())
	})

	t.Run("register fail", func(t *testing.T) {
		t.Run("panic", func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("should panic")
				}
			}()

			clpgxmigrate.Register(nil)
		})
	})
}
