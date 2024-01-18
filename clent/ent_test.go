package clent_test

import (
	"testing"

	"entgo.io/ent/dialect"
	"github.com/crewlinker/clgo/clent"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClent(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clent")
}

var _ = Describe("clid field", func() {
	It("should panic with invalid length", func() {
		Expect(func() {
			clent.Field.CLID("id", "a")
		}).To(PanicWith(MatchRegexp(`clid prefix not of length`)))
	})

	It("should have constructed correct field", func() {
		field := clent.Field.CLID("id2", "orga")

		Expect(field.Descriptor().Name).To(Equal(`id2`))
		Expect(field.Descriptor().SchemaType).To(Equal(map[string]string{
			dialect.Postgres: "varchar(31)",
		}))
	})

	It("should validate prefix", func() {
		Expect(clent.Validator.CLID("foo")("foo-01HMEFW25X3NAWEDGMPPYM1C6K")).To(Succeed())
		Expect(clent.Validator.CLID("foo")("bar-01HMEFW25X3NAWEDGMPPYM1C6K")).To(
			MatchError(ContainSubstring(`'bar' is invalid, expected: 'foo'`)))
	})

	It("should validate prefix", func() {
		Expect(clent.Validator.CLID("foo")("bar_01HMEFW25X3NAWEDGMPPYM1C6K")).To(
			MatchError(ContainSubstring(`no separator`)))
		Expect(clent.Validator.CLID("foo")("foo-ad")).To(
			MatchError(ContainSubstring(`invalid ulid`)))
	})
})
